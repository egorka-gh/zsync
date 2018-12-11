package repo

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"

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

	//run procedure
	var path = b.dbFolder + filename
	var sql = "CALL " + table + "_getcnv(?, ?, ?)"
	var newVersion int
	err := b.db.GetContext(ctx, &newVersion, sql, source, start, path)
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
	var path = b.dbFolder + pack.Pack
	var sql = "CALL " + pack.Table + "_runcnv(?, ?, ?)"
	_, err := b.db.ExecContext(ctx, sql, pack.Source, pack.End, path)
	return err

}

func (b *basicRepository) CalcLevels(ctx context.Context, balanceDate time.Time) error {
	return errors.New("Not implemented")
}

func (b *basicRepository) CalcBalance(ctx context.Context, balanceDate time.Time) error {
	return errors.New("Not implemented")
}

func (b *basicRepository) FixVersions(ctx context.Context, source string) error {
	_, err := b.db.ExecContext(ctx, "CALL fix_version(?)", source)
	return err
}

func (b *basicRepository) AddActivity(ctx context.Context, activity service.Activity) error {
	return errors.New("Not implemented")
}

func (b *basicRepository) GetLevel(ctx context.Context, card string) (int, error) {
	return 0, errors.New("Not implemented")
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
