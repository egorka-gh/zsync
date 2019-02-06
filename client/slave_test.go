package client

import (
	"os"
	"testing"

	endpoint "github.com/go-kit/kit/endpoint"
	log "github.com/go-kit/kit/log"
	http "github.com/go-kit/kit/transport/http"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	endpoint1 "github.com/egorka-gh/zbazar/zsync/pkg/endpoint"
	http1 "github.com/egorka-gh/zbazar/zsync/pkg/http"
	"github.com/egorka-gh/zbazar/zsync/pkg/repo"
	"github.com/egorka-gh/zbazar/zsync/pkg/service"
)

//var logger log.Logger

//NewDb creates new repro
func NewDb(cnn, folder string) (service.Repository, *sqlx.DB, error) {
	//"root:3411@tcp(127.0.0.1:3306)/pshdata"
	var db *sqlx.DB
	db, err := sqlx.Connect("mysql", cnn)
	if err != nil {
		return nil, nil, err
	}
	return repo.New(db, folder), db, nil
}

func initLoger(toFile string) log.Logger {
	var logger log.Logger
	if toFile != "" {
		logger = log.NewLogfmtLogger(&lumberjack.Logger{
			Filename:   toFile,
			MaxSize:    5, // megabytes
			MaxBackups: 3,
			MaxAge:     10, //days
		})
	} else {
		logger = log.NewLogfmtLogger(os.Stderr)
	}
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)
	logger.Log("tracer", "none")
	return logger
}

func defaultHttPOptions(logger log.Logger) map[string][]http.ServerOption {
	options := map[string][]http.ServerOption{
		"AddActivity": {http.ServerErrorEncoder(http1.ErrorEncoder), http.ServerErrorLogger(logger)},
		"GetLevel":    {http.ServerErrorEncoder(http1.ErrorEncoder), http.ServerErrorLogger(logger)},
		"ListVersion": {http.ServerErrorEncoder(http1.ErrorEncoder), http.ServerErrorLogger(logger)},
		"PackDone":    {http.ServerErrorEncoder(http1.ErrorEncoder), http.ServerErrorLogger(logger)},
		"PullPack":    {http.ServerErrorEncoder(http1.ErrorEncoder), http.ServerErrorLogger(logger)},
		"PushPack":    {http.ServerErrorEncoder(http1.ErrorEncoder), http.ServerErrorLogger(logger)},
	}
	return options
}

func startMasterService(done <-chan interface{}) error {
	const exchFolder string = "D:\\Buffer\\zexch\\00"
	logger := initLoger("D:\\Buffer\\zexch\\00\\log\\00.log")
	rep, _, err := NewDb("root:3411@tcp(127.0.0.1:3306)/pshdata", exchFolder)
	if err != nil {
		return err
	}
	//???
	defer rep.Close()
	var mw = []service.Middleware{}
	mw = append(mw, service.LoggingMiddleware(logger))
	svc := service.New(mw, rep, "00", exchFolder)
	var em map[string][]endpoint.Middleware
	eps := endpoint1.New(svc, em)

	//TODO remove
	if eps.AddActivity != nil {
		return nil
	}

	return nil
}

func TestSyncSlave(t *testing.T) {
	const exchFolder string = "D:\\Buffer\\zexch\\zs"
	logger := initLoger("D:\\Buffer\\zexch\\zs\\log\\zsync.log")

	var mrep service.Repository
	mrep, _, err := NewDb("root:3411@tcp(127.0.0.1:3306)/zslave", exchFolder)
	if err != nil {
		t.Fatal(err)
	}
	defer mrep.Close()

	//TODO remove
	logger.Log("remove", "me")
}
