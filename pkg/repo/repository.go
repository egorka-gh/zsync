package repo

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/egorka-gh/zbazar/zsync/pkg/service"
	//_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type basicRepository struct {
	db       *sqlx.DB
	dbFolder string
}

//ListSource gets source list exclude passed source (as a rule exclude self)
func (b *basicRepository) ListSource(ctx context.Context, source string) ([]service.Source, error) {
	var list []service.Source

	var ssql = "SELECT cs.id, cs.url FROM cnv_source cs WHERE cs.id != ?"
	err := b.db.SelectContext(ctx, &list, ssql, source)
	return list, err
}

func (b *basicRepository) ListVersion(ctx context.Context, source string) ([]service.Version, error) {
	var list []service.Version

	//var ssql = "SELECT source, table_name, latest_version FROM cnv_version WHERE source = ? ORDER BY syncorder"
	var sb strings.Builder
	sb.WriteString("SELECT ? source, table_name, MAX(latest_version) latest_version FROM cnv_version")
	sb.WriteString(" WHERE source IN (?, IF(? = '00', '000', ''))")
	sb.WriteString(" GROUP BY table_name, syncorder ORDER BY syncorder")
	var ssql = sb.String()
	err := b.db.SelectContext(ctx, &list, ssql, source, source, source)
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
	_, err := b.db.ExecContext(ctx, "CALL recalc_level(?)", balanceDate.Format("2006-01-02"))
	return err
}

func (b *basicRepository) CalcBalance(ctx context.Context, fromDate time.Time) error {
	_, err := b.db.ExecContext(ctx, "CALL recalc_balance(?)", fromDate.Format("2006-01-02"))
	return err
}

func (b *basicRepository) FixVersions(ctx context.Context, source string) error {
	_, err := b.db.ExecContext(ctx, "CALL fix_version(?)", source)
	return err
}

func (b *basicRepository) AddActivity(ctx context.Context, activity service.Activity) error {
	var sql = "INSERT INTO client_activity (source, doc_id, card, doc_date, doc_sum, bonuce_sum) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := b.db.ExecContext(ctx, sql, activity.Source, activity.Doc, activity.Card, activity.DocDate, activity.DocSum, activity.BonuceSum)
	return err
}

func (b *basicRepository) GetLevel(ctx context.Context, card string) (int, error) {
	var sql = "CALL get_level(?)"
	var level int
	err := b.db.GetContext(ctx, &level, sql, card)
	if err != nil {
		return 0, err
	}
	return level, nil
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

func (b *basicRepository) ExchangeFolder() string {
	return b.dbFolder
}

func (b *basicRepository) Close() {
	b.db.Close()
}

//NewTest creates new Repository, expect mysql sqlx.DB
func NewTest(connection, exchangeFolder string) (service.Repository, *sqlx.DB, error) {
	if !os.IsPathSeparator(exchangeFolder[len(exchangeFolder)-1]) {
		exchangeFolder = exchangeFolder + string(os.PathSeparator)
	}
	var db *sqlx.DB
	db, err := sqlx.Connect("mysql", connection)
	if err != nil {
		return nil, nil, err
	}

	return &basicRepository{
		db:       db,
		dbFolder: exchangeFolder,
	}, db, nil
}

//New creates new Repository
func New(connection, exchangeFolder string) (service.Repository, error) {
	/*
		if !os.IsPathSeparator(exchangeFolder[len(exchangeFolder)-1]) {
			exchangeFolder = exchangeFolder + string(os.PathSeparator)
		}
		var db *sqlx.DB
		db, err := sqlx.Connect("mysql", connection)
		if err != nil {
			return nil, err
		}
		return &basicRepository{
			db:       db,
			dbFolder: exchangeFolder,
		}, nil
	*/
	rep, _, err := NewTest(connection, exchangeFolder)
	return rep, err
}
