package service

import (
	"github.com/egorka-gh/zbazar/zsync/repo"
	"github.com/jmoiron/sqlx"
)

func newDb(cnn, folder string) (Repository, *sqlx.DB, error) {
	//"root:3411@tcp(127.0.0.1:3306)/pshdata"
	var db *sqlx.DB
	db, err := sqlx.Connect("mysql", cnn)
	if err != nil {
		return nil, nil, err
	}
	return repo.New(db, folder), db, nil
}
