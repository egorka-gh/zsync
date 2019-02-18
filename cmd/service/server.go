package service

import (
	endpoint "github.com/egorka-gh/zbazar/zsync/pkg/endpoint"
	"github.com/egorka-gh/zbazar/zsync/pkg/repo"
	service "github.com/egorka-gh/zbazar/zsync/pkg/service"
	log "github.com/go-kit/kit/log"
	"github.com/kardianos/osext"
	group "github.com/oklog/oklog/pkg/group"
	"github.com/spf13/viper"
	//lightsteptracergo "github.com/lightstep/lightstep-tracer-go"

	opentracinggo "github.com/opentracing/opentracing-go"
	//zipkingoopentracing "github.com/openzipkin/zipkin-go-opentracing"

	"os"
	//appdash "sourcegraph.com/sourcegraph/appdash"
	//opentracing "sourcegraph.com/sourcegraph/appdash/opentracing"
)

var tracer opentracinggo.Tracer
var logger log.Logger

// Define our flags. Your service probably won't need to bind listeners for
// all* supported transports, but we do it here for demonstration purposes.
//var fs = flag.NewFlagSet("zsync", flag.ExitOnError)
//var debugAddr = fs.String("debug.addr", ":8080", "Debug and metrics listen address")
//var httpAddr = fs.String("http-addr", ":8081", "HTTP listen address")
//var mysqlCnn = fs.String("mysql", "", "MySQL connection string")
//var exchangeFolder = fs.String("folder", "", "MySQL exchange folder")
//var logFolder = fs.String("log", "", "Log folder")
//var instanseID = fs.String("id", "", "Instanse ID")

//var grpcAddr = fs.String("grpc-addr", ":8082", "gRPC listen address")
//var thriftAddr = fs.String("thrift-addr", ":8083", "Thrift listen address")
//var thriftProtocol = fs.String("thrift-protocol", "binary", "binary, compact, json, simplejson")
//var thriftBuffer = fs.Int("thrift-buffer", 0, "0 for unbuffered")
//var thriftFramed = fs.Bool("thrift-framed", false, "true to enable framing")
//var zipkinURL = fs.String("zipkin-url", "", "Enable Zipkin tracing via a collector URL e.g. http://localhost:9411/api/v1/spans")
//var lightstepToken = fs.String("lightstep-token", "", "Enable LightStep tracing via a LightStep access token")
//var appdashAddr = fs.String("appdash-addr", "", "Enable Appdash tracing via an Appdash server host:port")

//RunServer strat service & client
func RunServer() {
	viper.SetDefault("http-addr", ":8081")  //HTTP listen addres
	viper.SetDefault("debug-addr", ":8080") //Debug and metrics listen address
	viper.SetDefault("mysql", "")           //MySQL connection string
	viper.SetDefault("folder", "")          //MySQL exchange folder
	viper.SetDefault("log", "")             //Log folder
	viper.SetDefault("id", "00")            //Instanse ID
	viper.SetDefault("master-url", "")      //master server url (need only for slave server)
	viper.SetDefault("sync-interval", "10") //sinc interval (minutes)
	viper.SetDefault("balance-hour", "2")   //hour of daily balance recalculation
	viper.SetDefault("lavel-days", "1,2")   //days of monthly level recalculation (comma sepparated)
	viper.SetDefault("lavel-hour", "2")     //hour of monthly level recalculation

	path, err := osext.ExecutableFolder()
	if err != nil {
		path = "."
	}
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	err = viper.ReadInConfig()
	/*
		if err != nil {
			logger.Info(err)
			logger.Info("Start using default setings")
		}
	*/

	fs.Parse(os.Args[1:])

	// Create a single logger, which we'll use and give to other components.
	initLoger(viper.GetString("log"))

	var debugAddr = viper.GetString("debug-addr")
	var httpAddr = viper.GetString("http-addr")
	var mysqlCnn = viper.GetString("mysql")
	var exchangeFolder = viper.GetString("folder")
	var logFolder = viper.GetString("log")
	var instanseID = viper.GetString("id")
	var masterURL = viper.GetString("master")

	if instanseID == "" {
		logger.Log("Error", "Instanse ID not set")
		return
	}
	if instanseID != "00" && masterURL == "" {
		logger.Log("Error", "Master server url not set")
		return
	}
	if mysqlCnn == "" {
		logger.Log("Error", "MySQL connection string not set")
		return
	}
	if exchangeFolder == "" {
		logger.Log("Error", "MySQL exchange folder not set")
		return
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
	rep, err = repo.New(mysqlCnn, exchangeFolder)
	if err != nil {
		logger.Log("Repository", "Connect", "err", err)
		return
	}
	defer rep.Close()

	svcLog := log.With(logger, "thread", "service")
	svc := service.New(getServiceMiddleware(svcLog), rep, instanseID, exchangeFolder)
	eps := endpoint.New(svc, getEndpointMiddleware(svcLog))
	g := createService(eps)
	initMetricsEndpoint(g)
	initCancelInterrupt(g)
	logger.Log("exit", g.Run())
}

func initClientScheduler(g *group.Group) {
}
