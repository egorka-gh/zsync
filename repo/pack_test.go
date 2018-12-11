package repo

import (
	"context"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/egorka-gh/zbazar/zsync/pkg/service"
)

func newDb(cnn, folder string) (service.Repository, error) {
	//"root:3411@tcp(127.0.0.1:3306)/pshdata"
	var db *sqlx.DB
	db, err := sqlx.Connect("mysql", cnn)
	if err != nil {
		return nil, err
	}
	return New(db, folder), nil
}

/*
func TestFixVersion(t *testing.T){
	mdb, err := newDb("root:3411@tcp(127.0.0.1:3306)/pshdata", "D:\\Buffer\\zexch")
	if err != nil {
		t.Fatal(err)
	}
}
*/

func TestGetVersion(t *testing.T) {
	var mdb service.Repository
	mdb, err := newDb("root:3411@tcp(127.0.0.1:3306)/pshdata", "D:\\Buffer\\zexch")
	if err != nil {
		t.Fatal(err)
	}

	ver, err := mdb.ListVersion(context.Background(), "00")
	if err != nil {
		t.Error(err)
	}
	t.Log(ver)

	err = mdb.FixVersions(context.Background(), "00")
	if err != nil {
		t.Error(err)
	}

	ver, err = mdb.ListVersion(context.Background(), "00")
	if err != nil {
		t.Error(err)
	}
	t.Log(ver)

}
