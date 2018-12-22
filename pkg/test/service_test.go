package test

import (
	"context"
	"os"
	"testing"

	log "github.com/go-kit/kit/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	"github.com/egorka-gh/zbazar/zsync/pkg/repo"
	"github.com/egorka-gh/zbazar/zsync/pkg/service"
)

var logger log.Logger

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

func initLoger(toFile bool) {
	if toFile {
		logger = log.NewLogfmtLogger(&lumberjack.Logger{
			Filename:   "D:\\Buffer\\zexch\\log\\zsync.log",
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

}

func TestFixVersionMaster(t *testing.T) {
	var exchFolder = "D:\\Buffer\\zexch"
	initLoger(true)

	var mrep service.Repository
	mrep, mdb, err := NewDb("root:3411@tcp(127.0.0.1:3306)/pshdata", exchFolder)
	if err != nil {
		t.Fatal(err)
	}
	defer mdb.Close()

	var mw = []service.Middleware{}
	mw = append(mw, service.LoggingMiddleware(logger))
	svc := service.New(mw, mrep, "00", exchFolder)

	ver0, err := svc.ListVersion(context.Background(), "")
	if err != nil {
		t.Error(err)
	}
	t.Log(ver0)

}
