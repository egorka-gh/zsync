package repo

import (
	"context"
	"os"
	"time"

	"github.com/egorka-gh/zbazar/zsync/pkg/service"
	"github.com/jmoiron/sqlx"
)

type basicRepository struct {
	db       *sqlx.DB
	dbFolder string
}

func (b *basicRepository) ListVersion(ctx context.Context, source string) ([]service.Version, error) {
	var list []service.Version
	var ssql = "SELECT source, table_name, latest_version FROM cnv_version WHERE source = ? ORDER BY syncorder"
	err := b.db.SelectContext(ctx, &list, ssql, source)
	return list, err
}

func (b *basicRepository) CreatePack(ctx context.Context, source, table, filename string, start int) (service.VersionPack, error) {
	var pack = service.VersionPack{
		Source: source,
		Table:  table,
		Start:  start,
		Pack:   filename,
	}

	//check delete pack file before sql
	err := b.DelPack(ctx, pack)
	if err != nil {
		pack.Pack = ""
		return pack, err
	}

	//run procwdure
	var path = b.dbFolder + filename
	var sql = "CALL " + table + "_getcnv(?, ?, ?)"
	var newVersion int
	err = b.db.GetContext(ctx, &newVersion, sql, source, start, path)
	if err != nil {
		pack.Pack = ""
		return pack, err
	}

	//check new version
	if newVersion <= start {
		pack.Pack = ""
	}
	pack.End = newVersion

	return pack, nil
}

func (b *basicRepository) ExecPack(ctx context.Context, pack service.VersionPack) error {

	if pack.Pack == "" || pack.End <= pack.Start {
		return nil
	}

}

func (b *basicRepository) DelPack(ctx context.Context, pack service.VersionPack) error {
	if pack.Pack != "" {
		var path = b.dbFolder + pack.Pack
		err := os.Remove(path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (b *basicRepository) CalcLevels(ctx context.Context, balanceDate time.Time) error {

}

func (b *basicRepository) CalcBalance(ctx context.Context, balanceDate time.Time) error {

}

func (b *basicRepository) FixVersions(ctx context.Context, source string) error {

}

func (b *basicRepository) AddActivity(ctx context.Context, activity service.Activity) error {

}

func (b *basicRepository) GetLevel(ctx context.Context, card string) (int, error) {

}

//New creates new Repository, expect mysql sqlx.DB
func New(rep *sqlx.DB, exchangeFolder string) service.Repository {
	/*
	   db, err = sqlx.Connect("mysql", viper.GetString("ConnectionString"))
	   	if err != nil {..
	*/
	if !os.IsPathSeparator(exchangeFolder[len(exchangeFolder)-1]) {
		exchangeFolder = exchangeFolder + string(os.PathSeparator)
	}

	return &basicRepository{
		db:       rep,
		dbFolder: exchangeFolder,
	}
}
