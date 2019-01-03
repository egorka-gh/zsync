package test

import (
	"context"
	"os"
	"testing"
	"time"

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

func TestMoveVersionMaster(t *testing.T) {
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

	//some db activity
	err = updateMaster(mdb)
	if err != nil {
		t.Error(err)
	}

	//in client task
	err = mrep.FixVersions(context.Background(), "00")
	if err != nil {
		t.Error(err)
	}

	ver1, err := svc.ListVersion(context.Background(), "")
	if err != nil {
		t.Error(err)
	}
	t.Log(ver1)

	//check if version updated
	for _, v0 := range ver0 {
		for _, v := range ver1 {
			if v0.Table == v.Table && (v0.Version+1) != v.Version {
				t.Error(v0.Table, " Expected version ", (v0.Version + 1), ", got ", v.Version)
			}
		}
	}

}

func TestMoveVersionSlave(t *testing.T) {
	var exchFolder = "D:\\Buffer\\zexch"
	initLoger(true)

	var mrep service.Repository
	mrep, mdb, err := NewDb("root:3411@tcp(127.0.0.1:3306)/zslave", exchFolder)
	if err != nil {
		t.Fatal(err)
	}
	defer mdb.Close()

	var mw = []service.Middleware{}
	mw = append(mw, service.LoggingMiddleware(logger))
	svc := service.New(mw, mrep, "zs", exchFolder)

	ver0, err := svc.ListVersion(context.Background(), "")
	if err != nil {
		t.Error(err)
	}
	t.Log(ver0)

	//some db activity
	err = updateSlave(mdb)
	if err != nil {
		t.Error(err)
	}

	//in client task
	err = mrep.FixVersions(context.Background(), "zs")
	if err != nil {
		t.Error(err)
	}

	ver1, err := svc.ListVersion(context.Background(), "")
	if err != nil {
		t.Error(err)
	}
	t.Log(ver1)

	//check if version updated
	for _, v0 := range ver0 {
		for _, v := range ver1 {
			if v0.Table == v.Table && (v0.Version+1) != v.Version {
				t.Error(v0.Table, " Expected version ", (v0.Version + 1), ", got ", v.Version)
			}
		}
	}

}

func TestPullPack(t *testing.T) {
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

	//create pack for first table
	v := ver0[0]
	t.Log(v)

	//same version (no pack created)
	p, err := svc.PullPack(context.Background(), "zs", v.Table, v.Version)
	if err != nil {
		t.Error(err)
	}
	t.Log(p)

	if p.Pack != "" {
		t.Error("Pack not expected ", p.Pack, ". Version in db ", v.Version, ", asked version ", p.Start)
	}

	//older version (pack created)
	p, err = svc.PullPack(context.Background(), "zs", v.Table, v.Version-1)
	if err != nil {
		t.Error(err)
	}
	t.Log(p)

	if p.Pack == "" {
		t.Error("Pack not created. Version in db ", v.Version, ", asked version ", p.Start)
	}

	//remove pack
	err = svc.PackDone(context.Background(), p)
	if err != nil {
		t.Error(err)
	}

}

func TestAddActivity(t *testing.T) {
	var exchFolder = "D:\\Buffer\\zexch"
	initLoger(true)

	var mrep service.Repository
	mrep, mdb, err := NewDb("root:3411@tcp(127.0.0.1:3306)/zslave?parseTime=true", exchFolder)
	if err != nil {
		t.Fatal(err)
	}
	defer mdb.Close()

	var mw = []service.Middleware{}
	mw = append(mw, service.LoggingMiddleware(logger))
	svc := service.New(mw, mrep, "zs", exchFolder)

	dt := time.Now()

	var a = service.Activity{
		Doc:       dt.Format(time.RFC3339),
		Card:      "100006",
		DocDate:   dt.Format("2006-01-02 15:04:05"),
		DocSum:    101.0,
		BonuceSum: 100.0,
	}

	err = svc.AddActivity(context.Background(), a)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(a)

	a2, err := loadActivity(mdb, "zs", a.Doc)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(a2)
	if a.Card != a2.Card || a.DocDate != a2.DocDate || a.DocSum != a2.DocSum || a.BonuceSum != a2.BonuceSum {
		t.Error("Activity mismatch")
	}
}

//TODO TestGetLevel
