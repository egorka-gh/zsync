package service

import (
	"errors"
	"fmt"
	"net"
	http1 "net/http"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/egorka-gh/zbazar/zsync/client"
	endpoint "github.com/egorka-gh/zbazar/zsync/pkg/endpoint"
	"github.com/egorka-gh/zbazar/zsync/pkg/http"
	"github.com/egorka-gh/zbazar/zsync/pkg/repo"
	"github.com/egorka-gh/zbazar/zsync/pkg/scheduler"
	service "github.com/egorka-gh/zbazar/zsync/pkg/service"
	endpoint1 "github.com/go-kit/kit/endpoint"
	log "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/kardianos/osext"
	group "github.com/oklog/oklog/pkg/group"
	prometheus1 "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
	//lightsteptracergo "github.com/lightstep/lightstep-tracer-go"

	opentracinggo "github.com/opentracing/opentracing-go"
	//zipkingoopentracing "github.com/openzipkin/zipkin-go-opentracing"

	"os"
	//appdash "sourcegraph.com/sourcegraph/appdash"
	//opentracing "sourcegraph.com/sourcegraph/appdash/opentracing"
)

var tracer opentracinggo.Tracer
var logger log.Logger

//ReadConfig init/read viper config
func ReadConfig() error {
	viper.SetDefault("http-addr", ":8081")  //HTTP listen addres
	viper.SetDefault("debug-addr", ":8080") //Debug and metrics listen address
	viper.SetDefault("mysql", "")           //MySQL connection string
	viper.SetDefault("folder", "")          //MySQL exchange folder
	viper.SetDefault("log", "")             //Log folder
	viper.SetDefault("id", "")              //Instanse ID
	viper.SetDefault("master-url", "")      //master server url (need only for slave server)
	viper.SetDefault("sync-interval", "10") //sinc interval (minutes)
	viper.SetDefault("balance-hour", "2")   //hour of daily balance recalculation
	viper.SetDefault("level-days", "1,2")   //days of monthly level recalculation (comma sepparated)
	viper.SetDefault("level-hour", "2")     //hour of monthly level recalculation

	path, err := osext.ExecutableFolder()
	if err != nil {
		path = "."
	}
	//fmt.Println("Path ", path)
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	return viper.ReadInConfig()
	/*
		if err != nil {
			logger.Info(err)
			logger.Info("Start using default setings")
		}
	*/
}

//InitServerGroup creates server run group, vs out CancelInterrupt runner
func InitServerGroup() (*group.Group, service.Repository, error) {

	// Create a single logger, which we'll use and give to other components.
	logger = initLoger(viper.GetString("log"))

	//var debugAddr = viper.GetString("debug-addr")
	//var httpAddr = viper.GetString("http-addr")
	var mysqlCnn = viper.GetString("mysql")
	var exchangeFolder = viper.GetString("folder")
	//var logFolder = viper.GetString("log")
	var instanseID = viper.GetString("id")
	var masterURL = viper.GetString("master-url")

	if instanseID == "" {
		return nil, nil, errors.New("Instanse ID not set")
	}
	if instanseID != "00" && masterURL == "" {
		return nil, nil, errors.New("Master server url not set")
	}
	if mysqlCnn == "" {
		return nil, nil, errors.New("MySQL connection string not set")
	}
	if exchangeFolder == "" {
		return nil, nil, errors.New("MySQL exchange folder not set")
	}

	logger.Log("InstanseID", instanseID)
	logger.Log("MySQLConnection", mysqlCnn)
	logger.Log("ExchangeFolder", exchangeFolder)
	if instanseID != "00" {
		logger.Log("MasterURL", masterURL)
	}
	logger.Log("tracer", "none")
	tracer = opentracinggo.GlobalTracer()

	var rep service.Repository
	rep, err := repo.New(mysqlCnn, exchangeFolder)
	if err != nil {
		logger.Log("Repository", "Connect", "err", err)
		return nil, nil, err
	}
	//defer rep.Close()

	//create client
	var cli *client.Client
	if instanseID != "00" {
		//start slave
		cli = client.NewSlave(rep, instanseID, masterURL, logger)
	} else {
		//start master
		cli = client.NewMaster(rep, instanseID, logger)
	}

	//create service
	svcLog := log.With(logger, "thread", "service")
	svc := service.New(getServiceMiddleware(svcLog), rep, instanseID, exchangeFolder)
	eps := endpoint.New(svc, getEndpointMiddleware(svcLog))

	//run group
	g := createService(eps)
	initClientScheduler(g, cli)
	initMetricsEndpoint(g)
	return g, rep, nil
}

//RunServer strat service & client vs os cli (listen os Signal CancelInterrupt)
func RunServer() {
	g, rep, err := InitServerGroup()
	if err != nil {
		if logger != nil {
			logger.Log("Error", err)
		}
		return
	}
	defer rep.Close()
	initCancelInterrupt(g)
	logger.Log("exit", g.Run())
}

func initHttpHandler(endpoints endpoint.Endpoints, g *group.Group) {
	options := defaultHttpOptions(logger, tracer)
	// Add your http options here

	httpHandler := http.NewHTTPHandler(endpoints, options)
	var httpAddr = viper.GetString("http-addr")
	var exchangeFolder = viper.GetString("folder")

	m, ok := httpHandler.(*http1.ServeMux)
	if ok {
		logger.Log("transport", "HTTP", "serve", exchangeFolder, "addr", httpAddr+http.PackPattern)
		fs := http1.FileServer(http1.Dir(exchangeFolder))
		fs = http.LoggingStatusHandler(fs, logger)
		m.Handle(http.PackPattern, http1.StripPrefix(http.PackPattern, fs))
	} else {
		logger.Log("transport", "HTTP", "during", "Handle "+http.PackPattern, "err", "Can't get ServeMux")
	}

	httpListener, err := net.Listen("tcp", httpAddr)
	if err != nil {
		logger.Log("transport", "HTTP", "during", "Listen", "err", err)
	}
	g.Add(func() error {
		logger.Log("transport", "HTTP", "addr", httpAddr)
		return http1.Serve(httpListener, httpHandler)
	}, func(error) {
		httpListener.Close()
	})

}
func getServiceMiddleware(logger log.Logger) (mw []service.Middleware) {
	mw = []service.Middleware{}
	mw = addDefaultServiceMiddleware(logger, mw)
	// Append your middleware here

	return
}
func getEndpointMiddleware(logger log.Logger) (mw map[string][]endpoint1.Middleware) {
	mw = map[string][]endpoint1.Middleware{}
	duration := prometheus.NewSummaryFrom(prometheus1.SummaryOpts{
		Help:      "Request duration in seconds.",
		Name:      "request_duration_seconds",
		Namespace: "example",
		Subsystem: "zsync",
	}, []string{"method", "success"})
	addDefaultEndpointMiddleware(logger, duration, mw)
	// Add you endpoint middleware here

	return
}
func initMetricsEndpoint(g *group.Group) {
	var debugAddr = viper.GetString("debug-addr")
	http1.DefaultServeMux.Handle("/metrics", promhttp.Handler())
	debugListener, err := net.Listen("tcp", debugAddr)
	if err != nil {
		logger.Log("transport", "debug/HTTP", "during", "Listen", "err", err)
	}
	g.Add(func() error {
		logger.Log("transport", "debug/HTTP", "addr", debugAddr)
		return http1.Serve(debugListener, http1.DefaultServeMux)
	}, func(error) {
		debugListener.Close()
	})
}
func initCancelInterrupt(g *group.Group) {
	cancelInterrupt := make(chan struct{})
	g.Add(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-c:
			return fmt.Errorf("received signal %s", sig)
		case <-cancelInterrupt:
			return nil
		}
	}, func(error) {
		close(cancelInterrupt)
	})
}

func initLoger(logPath string) log.Logger {
	var logger log.Logger
	if logPath == "" {
		logger = log.NewLogfmtLogger(os.Stderr)
	} else {
		path := logPath
		if !os.IsPathSeparator(path[len(path)-1]) {
			path = path + string(os.PathSeparator)
		}
		path = path + "zsync.log"
		logger = log.NewLogfmtLogger(&lumberjack.Logger{
			Filename:   path,
			MaxSize:    5, // megabytes
			MaxBackups: 5,
			MaxAge:     60, //days
		})
	}
	logger = log.With(logger, "ts", log.DefaultTimestamp) // .DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	return logger
}

func initClientScheduler(g *group.Group, cli *client.Client) {

	interval := viper.GetInt("sync-interval")
	if interval < 10 {
		interval = 10
	}
	balanceHour := viper.GetInt("balance-hour")
	if balanceHour <= 0 {
		balanceHour = 2
	}
	levelHour := viper.GetInt("level-hour")
	if levelHour <= 0 {
		levelHour = 2
	}

	levelDaysStr := viper.GetString("level-days")
	var levelDays []int
	a := strings.Split(levelDaysStr, ",")
	for _, s := range a {
		i, err := strconv.Atoi(s)
		if err == nil && i > 0 && i < 31 {
			levelDays = append(levelDays, i)
		}
	}
	if len(levelDays) == 0 {
		levelDays = append(levelDays, 1)
	}

	scheduler := scheduler.New()

	scheduler.AddPeriodic(time.Duration(interval)*time.Minute, cli.Sync)
	if viper.GetString("id") == "00" {
		scheduler.AddDaily(balanceHour, cli.CalcBalance)
		for _, d := range levelDays {
			scheduler.AddMonthly(d, levelHour, cli.CalcLevels)
		}
	} else {
		for _, d := range levelDays {
			scheduler.AddMonthly(d, levelHour, cli.CleanUp)
		}
	}
	scheduler.AddPeriodic(time.Duration(interval)*time.Minute, cli.FixVersions)

	g.Add(func() error {
		logger.Log("client", "scheduler", "sync", interval, "balanceHour", balanceHour, "levelDays", levelDaysStr, "levelHour", levelHour)
		return scheduler.Run()
	}, func(error) {
		scheduler.Stop()
	})

}
