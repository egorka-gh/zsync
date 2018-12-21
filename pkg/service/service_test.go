package service

import (
	"testing"

	"github.com/egorka-gh/zbazar/zsync/pkg/repo"
	"github.com/jmoiron/sqlx"
)

//NewDb creates new repro
func NewDb(cnn, folder string) (Repository, *sqlx.DB, error) {
	//"root:3411@tcp(127.0.0.1:3306)/pshdata"
	var db *sqlx.DB
	db, err := sqlx.Connect("mysql", cnn)
	if err != nil {
		return nil, nil, err
	}
	return repo.New(db, folder), db, nil
}

func TestFixVersionMaster(t *testing.T) {
	var mrep Repository
	mrep, mdb, err := test.NewDb("root:3411@tcp(127.0.0.1:3306)/pshdata", "D:\\Buffer\\zexch")
	if err != nil {
		t.Fatal(err)
	}
	defer mdb.Close()
}
