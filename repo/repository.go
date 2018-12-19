package repo

import (
	"context"
	"os"
	"strings"
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

	//del pack file  if exists
	err := b.delPack(ctx, filename)
	if err != nil {
		pack.Pack = ""
		return pack, err
	}

	//run procedure
	var path = b.dbFolder + filename
	path = strings.Replace(path, "\\", "/", -1)
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
	var path = b.dbFolder + pack.Pack
	path = strings.Replace(path, "\\", "/", -1)
	//var sql = "CALL " + pack.Table + "_runcnv(?, ?, ?)"
	//_, err := b.db.ExecContext(ctx, sql, pack.Source, pack.End, path)
	var sb strings.Builder
	sb.WriteString("LOAD DATA INFILE '")
	sb.WriteString(path)
	sb.WriteString("' REPLACE  INTO TABLE ")
	sb.WriteString(pack.Table)
	var sql = sb.String()
	_, err := b.db.ExecContext(ctx, sql)
	if err == nil {
		sql = "UPDATE cnv_version SET latest_version = ? WHERE source = ? AND table_name = ?"
		result, err := b.db.ExecContext(ctx, sql, pack.End, pack.Source, pack.Table)
		//create version record if not exists
		rows, _ := result.RowsAffected()
		if err == nil && rows == 0 {
			sb.Reset()
			sb.WriteString("INSERT INTO cnv_version (source, table_name, latest_version, syncorder)")
			sb.WriteString(" SELECT DISTINCT ?, table_name, ?, syncorder")
			sb.WriteString(" FROM cnv_version")
			sb.WriteString(" WHERE table_name = ?")
			sb.WriteString(" ON DUPLICATE KEY UPDATE latest_version = ?")
			sql = sb.String()
			_, err = b.db.ExecContext(ctx, sql, pack.Source, pack.End, pack.Table, pack.End)
		}
	}
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

func (b *basicRepository) delPack(ctx context.Context, fileName string) (e0 error) {
	if fileName != "" {
		var path = b.dbFolder + fileName
		err := os.Remove(path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
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
