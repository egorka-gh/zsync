package client

import (
	"context"
	"encoding/json"
	"io/ioutil"
	http2 "net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

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
	"github.com/egorka-gh/zbazar/zsync/pkg/test"
)

//var logger log.Logger

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

func startClient(id, cnn, folder, log, mainURL string) (*Client, service.Repository, *sqlx.DB, error) {
	logger := initLoger(log)
	rep, db, err := repo.NewTest(cnn, folder)
	if err != nil {
		return nil, nil, nil, err
	}

	var c *Client
	if id != "00" {
		//start subordinate
		c = NewSubordinate(rep, id, mainURL, logger)
	} else {
		//start main
		c = NewMain(rep, id, logger)
	}

	return c, rep, db, nil
}

func startService(id, cnn, folder, log string) (*httptest.Server, service.Repository, *sqlx.DB, error) {
	logger := initLoger(log)

	var rep service.Repository
	rep, db, err := repo.NewTest(cnn, folder)
	if err != nil {
		return nil, nil, nil, err
	}

	var mw = []service.Middleware{}
	mw = append(mw, service.LoggingMiddleware(logger))
	svc := service.New(mw, rep, id, folder)
	var em map[string][]endpoint.Middleware
	eps := endpoint1.New(svc, em)

	httpHandler := http1.NewHTTPHandler(eps, defaultHttPOptions(logger))
	m, ok := httpHandler.(*http2.ServeMux)
	if ok {
		logger.Log("transport", "HTTP", "serve", folder, "addr", http1.PackPattern)
		fs := http2.FileServer(http2.Dir(folder))
		fs = http1.LoggingStatusHandler(fs, logger)
		m.Handle(http1.PackPattern, http2.StripPrefix(http1.PackPattern, fs))
	} else {
		logger.Log("transport", "HTTP", "during", "Handle "+http1.PackPattern, "err", "Can't get ServeMux")
	}

	srv := httptest.NewServer(httpHandler)
	return srv, rep, db, nil
}

func TestSubordinateAddActivity(t *testing.T) {
	srv, _, db, err := startService("zs", "root:3411@tcp(127.0.0.1:3306)/zsubordinate", "D:\\Buffer\\zexch\\zs", "D:\\Buffer\\zexch\\zs\\log\\zsync.log")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer srv.Close()

	//delete today activiti
	dt := time.Now()
	//dt.Format("2006-01-02 15:04:05")
	sql := "DELETE FROM client_activity WHERE source = 'zs' AND doc_id Like '" + dt.Format("2006-01-02") + "%'"
	_, err = db.Exec(sql)
	if err != nil {
		t.Log("Db err while delete", err)
	}

	//add some activity
	a1 := service.Activity{
		Source:    "zs", //4 test check, service will set its own ID
		Doc:       dt.Format("2006-01-02") + "_1",
		DocDate:   dt.Format("2006-01-02 15:04:05"),
		Card:      "100024",
		DocSum:    10.7,
		BonuceSum: 10.7,
	}
	a2 := service.Activity{
		Source:    "zs", //4 test,
		Doc:       dt.Format("2006-01-02") + "_2",
		DocDate:   dt.Format("2006-01-02 15:04:05"),
		Card:      "100024",
		DocSum:    10.7,
		BonuceSum: 10.7,
	}
	//{"activity":{"doc_id":"2019-02-14_1","card":"100024","doc_date":"2019-02-14","doc_sum":10.7,"bonuce_sum":10.7}}
	b1, _ := json.Marshal(endpoint1.AddActivityRequest{Activity: a1})
	b2, _ := json.Marshal(endpoint1.AddActivityRequest{Activity: a2})

	for _, testcase := range []struct {
		method string
		url    string
		body   string
		want   service.Activity
	}{
		{"POST", srv.URL + "/add-activity", string(b1), a1},
		{"POST", srv.URL + "/add-activity", string(b2), a2},
	} {
		t.Log("Request ", testcase.body)
		req, _ := http2.NewRequest(testcase.method, testcase.url, strings.NewReader(testcase.body))
		resp, _ := http2.DefaultClient.Do(req)
		body, _ := ioutil.ReadAll(resp.Body)
		t.Log("Responce ", string(body))
		a, err := test.LoadActivity(db, testcase.want.Source, testcase.want.Doc)
		if err != nil {
			t.Error("Db err ", err)
		}
		if testcase.want.Source != a.Source || testcase.want.Doc != a.Doc {
			t.Errorf("want %s; %s, got %s; %s", testcase.want.Source, testcase.want.Doc, a.Source, a.Doc)
		}
	}
	/**/
}

func TestSubordinateSync(t *testing.T) {
	/*
		//start subordinate
		subordinate, _, sdb, err := startService("zs", "root:3411@tcp(127.0.0.1:3306)/zsubordinate", "D:\\Buffer\\zexch\\zs", "D:\\Buffer\\zexch\\zs\\log\\zsync.log")
		if err != nil {
			t.Fatal(err)
		}
		defer sdb.Close()
		defer subordinate.Close()
		t.Log("Subordinate url ", subordinate.URL)
	*/

	//start main
	main, mrep, mdb, err := startService("00", "root:3411@tcp(127.0.0.1:3306)/pshdata", "D:\\Buffer\\zexch\\00", "D:\\Buffer\\zexch\\00\\log\\00.log")
	if err != nil {
		t.Fatal(err)
	}
	defer mdb.Close()
	defer main.Close()
	t.Log("Main url ", main.URL)
	mrep.FixVersions(context.Background(), "00")

	c, crep, cdb, err := startClient("zs", "root:3411@tcp(127.0.0.1:3306)/zsubordinate", "D:\\Buffer\\zexch\\zs", "D:\\Buffer\\zexch\\zs\\log\\zsync.log", main.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer crep.Close()

	sql := "UPDATE cnv_version SET latest_version = 0 WHERE source = '00'"
	_, err = cdb.Exec(sql)
	if err != nil {
		t.Fatal(err)
	}

	c.Sync(context.Background())

	v00, err := mrep.ListVersion(context.Background(), "00")
	if err != nil {
		t.Fatal(err)
	}
	vc, err := crep.ListVersion(context.Background(), "00")
	if err != nil {
		t.Fatal(err)
	}
	//check if version updated
	for _, v0 := range v00 {
		for _, v := range vc {
			if v0.Table == v.Table && v0.Version != v.Version {
				t.Error(v0.Table, " Expected remote version ", v0.Version, ", got ", v.Version)
			}
		}
	}

}

func TestMainSync(t *testing.T) {

	//start subordinate
	//main, mrep, mdb, err := startService("00", "root:3411@tcp(127.0.0.1:3306)/pshdata", "D:\\Buffer\\zexch\\00", "D:\\Buffer\\zexch\\00\\log\\00.log")
	main, mrep, mdb, err := startService("zs", "root:3411@tcp(127.0.0.1:3306)/zsubordinate", "D:\\Buffer\\zexch\\zs", "D:\\Buffer\\zexch\\zs\\log\\zsync.log")
	if err != nil {
		t.Fatal(err)
	}
	defer mdb.Close()
	defer main.Close()
	t.Log("Main url ", main.URL)
	mrep.FixVersions(context.Background(), "zs")

	//c, crep, cdb, err := startClient("zs", "root:3411@tcp(127.0.0.1:3306)/zsubordinate", "D:\\Buffer\\zexch\\zs", "D:\\Buffer\\zexch\\zs\\log\\zsync.log", main.URL)
	c, crep, cdb, err := startClient("00", "root:3411@tcp(127.0.0.1:3306)/pshdata", "D:\\Buffer\\zexch\\00", "D:\\Buffer\\zexch\\00\\log\\00.log", "")
	if err != nil {
		t.Fatal(err)
	}
	defer crep.Close()

	sql := "UPDATE cnv_source SET url = '" + main.URL + "' WHERE id = 'zs'"
	_, err = cdb.Exec(sql)
	if err != nil {
		t.Fatal(err)
	}

	sql = "UPDATE cnv_version SET latest_version = 0 WHERE source = 'zs'"
	_, err = cdb.Exec(sql)
	if err != nil {
		t.Fatal(err)
	}

	c.Sync(context.Background())

	v00, err := mrep.ListVersion(context.Background(), "zs")
	if err != nil {
		t.Fatal(err)
	}
	vc, err := crep.ListVersion(context.Background(), "zs")
	if err != nil {
		t.Fatal(err)
	}
	//check if version updated
	for _, v0 := range v00 {
		for _, v := range vc {
			if v0.Table == v.Table && v0.Version != v.Version {
				t.Error(v0.Table, " Expected remote version ", v0.Version, ", got ", v.Version)
			}
		}
	}

}
