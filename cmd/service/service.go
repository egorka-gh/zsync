package service

import (
	"flag"
	"fmt"

	endpoint "github.com/egorka-gh/zbazar/zsync/pkg/endpoint"
	http "github.com/egorka-gh/zbazar/zsync/pkg/http"
	"github.com/egorka-gh/zbazar/zsync/pkg/repo"
	service "github.com/egorka-gh/zbazar/zsync/pkg/service"
	endpoint1 "github.com/go-kit/kit/endpoint"
	log "github.com/go-kit/kit/log"
	prometheus "github.com/go-kit/kit/metrics/prometheus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
	//lightsteptracergo "github.com/lightstep/lightstep-tracer-go"
	group "github.com/oklog/oklog/pkg/group"
	opentracinggo "github.com/opentracing/opentracing-go"
	//zipkingoopentracing "github.com/openzipkin/zipkin-go-opentracing"
	"net"
	http1 "net/http"
	"os"
	"os/signal"

	prometheus1 "github.com/prometheus/client_golang/prometheus"
	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
	//appdash "sourcegraph.com/sourcegraph/appdash"
	//opentracing "sourcegraph.com/sourcegraph/appdash/opentracing"
	"syscall"
)

var tracer opentracinggo.Tracer
var logger log.Logger

// Define our flags. Your service probably won't need to bind listeners for
// all* supported transports, but we do it here for demonstration purposes.
var fs = flag.NewFlagSet("zsync", flag.ExitOnError)
var debugAddr = fs.String("debug.addr", ":8080", "Debug and metrics listen address")
var httpAddr = fs.String("http-addr", ":8081", "HTTP listen address")
var mysqlCnn = fs.String("mysql", "", "MySQL connection string")
var exchangeFolder = fs.String("folder", "", "MySQL exchange folder")
var logFolder = fs.String("log", "", "Log folder")
var instanseID = fs.String("id", "", "Instanse ID")

//var grpcAddr = fs.String("grpc-addr", ":8082", "gRPC listen address")
//var thriftAddr = fs.String("thrift-addr", ":8083", "Thrift listen address")
//var thriftProtocol = fs.String("thrift-protocol", "binary", "binary, compact, json, simplejson")
//var thriftBuffer = fs.Int("thrift-buffer", 0, "0 for unbuffered")
//var thriftFramed = fs.Bool("thrift-framed", false, "true to enable framing")
//var zipkinURL = fs.String("zipkin-url", "", "Enable Zipkin tracing via a collector URL e.g. http://localhost:9411/api/v1/spans")
//var lightstepToken = fs.String("lightstep-token", "", "Enable LightStep tracing via a LightStep access token")
//var appdashAddr = fs.String("appdash-addr", "", "Enable Appdash tracing via an Appdash server host:port")

//Run strart service
func Run() {
	fs.Parse(os.Args[1:])

	// Create a single logger, which we'll use and give to other components.
	if *logFolder == "" {
		logger = log.NewLogfmtLogger(os.Stderr)
	} else {
		path := *logFolder
		if !os.IsPathSeparator(path[len(path)-1]) {
			path = path + string(os.PathSeparator)
		}
		path = path + "zsync.log"
		logger = log.NewLogfmtLogger(&lumberjack.Logger{
			Filename:   path,
			MaxSize:    5, // megabytes
			MaxBackups: 3,
			MaxAge:     10, //days
		})
	}
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	if *instanseID == "" {
		logger.Log("Error", "Instanse ID not set")
		fs.Usage()
		return
	}
	if *mysqlCnn == "" {
		logger.Log("Error", "MySQL connection string not set")
		fs.Usage()
		return
	}
	if *exchangeFolder == "" {
		logger.Log("Error", "MySQL exchange folder not set")
		fs.Usage()
		return
	}
	logger.Log("InstanseID", *instanseID)
	logger.Log("MySQLConnection", *mysqlCnn)
	logger.Log("ExchangeFolder", *exchangeFolder)

	//  Determine which tracer to use. We'll pass the tracer to all the
	// components that use it, as a dependency
	/*
		if *zipkinURL != "" {
			logger.Log("tracer", "Zipkin", "URL", *zipkinURL)
			collector, err := zipkingoopentracing.NewHTTPCollector(*zipkinURL)
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
			defer collector.Close()
			recorder := zipkingoopentracing.NewRecorder(collector, false, "localhost:80", "zsync")
			tracer, err = zipkingoopentracing.NewTracer(recorder)
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
		} else if *lightstepToken != "" {
			logger.Log("tracer", "LightStep")
			tracer = lightsteptracergo.NewTracer(lightsteptracergo.Options{AccessToken: *lightstepToken})
			defer lightsteptracergo.FlushLightStepTracer(tracer)
		} else if *appdashAddr != "" {
			logger.Log("tracer", "Appdash", "addr", *appdashAddr)
			collector := appdash.NewRemoteCollector(*appdashAddr)
			tracer = opentracing.NewTracer(collector)
			defer collector.Close()
		} else {
			logger.Log("tracer", "none")
			tracer = opentracinggo.GlobalTracer()
		}
	*/

	logger.Log("tracer", "none")
	tracer = opentracinggo.GlobalTracer()

	var rep service.Repository
	rep, err := repo.New(*mysqlCnn, *exchangeFolder)
	if err != nil {
		logger.Log("Repository", "Connect", "err", err)
		return
	}
	defer rep.Close()

	svcLog := log.With(logger, "thread", "service")
	svc := service.New(getServiceMiddleware(svcLog), rep, *instanseID, *exchangeFolder)
	eps := endpoint.New(svc, getEndpointMiddleware(svcLog))
	g := createService(eps)
	initMetricsEndpoint(g)
	initCancelInterrupt(g)
	logger.Log("exit", g.Run())

}
func initHttpHandler(endpoints endpoint.Endpoints, g *group.Group) {
	options := defaultHttpOptions(logger, tracer)
	// Add your http options here

	httpHandler := http.NewHTTPHandler(endpoints, options)
	m, ok := httpHandler.(*http1.ServeMux)
	if ok {
		logger.Log("transport", "HTTP", "serve", *exchangeFolder, "addr", *httpAddr+http.PackPattern)
		fs := http1.FileServer(http1.Dir(*exchangeFolder))
		fs = http.LoggingStatusHandler(fs, logger)
		m.Handle(http.PackPattern, http1.StripPrefix(http.PackPattern, fs))
	} else {
		logger.Log("transport", "HTTP", "during", "Handle "+http.PackPattern, "err", "Can't get ServeMux")
	}

	httpListener, err := net.Listen("tcp", *httpAddr)
	if err != nil {
		logger.Log("transport", "HTTP", "during", "Listen", "err", err)
	}
	g.Add(func() error {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
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
	http1.DefaultServeMux.Handle("/metrics", promhttp.Handler())
	debugListener, err := net.Listen("tcp", *debugAddr)
	if err != nil {
		logger.Log("transport", "debug/HTTP", "during", "Listen", "err", err)
	}
	g.Add(func() error {
		logger.Log("transport", "debug/HTTP", "addr", *debugAddr)
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

func initLoger(toFile string) log.Logger {
	var logger log.Logger
	if toFile == "" {
		logger = log.NewLogfmtLogger(os.Stderr)
	} else {
		path := *logFolder
		if !os.IsPathSeparator(path[len(path)-1]) {
			path = path + string(os.PathSeparator)
		}
		path = path + "zsync.log"
		logger = log.NewLogfmtLogger(&lumberjack.Logger{
			Filename:   path,
			MaxSize:    5, // megabytes
			MaxBackups: 3,
			MaxAge:     10, //days
		})
	}
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	return logger
}
