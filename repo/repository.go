package repo

import (
	"context"
	"os"
	"time"

	"github.com/egorka-gh/zbazar/zsync1st/dto"

	"github.com/jmoiron/sqlx"
)

type basicRepository struct {
	db       *sqlx.DB
	dbFolder string
}

func (b *basicRepository) ListVersion(ctx context.Context, source string) ([]dto.Version, error) {
	var list []dto.Version
	var ssql = "SELECT source, table_name, latest_version FROM cnv_version WHERE source = ? ORDER BY syncorder"
	err := b.db.Select(&list, ssql, source)
	return list, err
}

func (b *basicRepository) CreatePack(ctx context.Context, source, table, filename string, start int) (dto.VersionPack, error) {
	//TODO check delete pack file before sql
	var pack = dto.VersionPack{
		Source: source,
		Table:  table,
		Start:  start,
		Pack:   filename,
	}

	var path = b.dbFolder + filename
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return pack, err
	}

	return pack, nil
}

func (b *basicRepository) ExecPack(ctx context.Context, pack dto.VersionPack) error {

}

func (b *basicRepository) DelPack(ctx context.Context, pack dto.VersionPack) error {

}

func (b *basicRepository) CalcLevels(ctx context.Context, balanceDate time.Time) error {

}

func (b *basicRepository) CalcBalance(ctx context.Context, balanceDate time.Time) error {

}

func (b *basicRepository) FixVersions(ctx context.Context, source string) error {

}

func (b *basicRepository) AddActivity(ctx context.Context, activity dto.Activity) error {

}

func (b *basicRepository) GetLevel(ctx context.Context, card string) (int, error) {

}

//New creates new Repository, expect mysql sqlx.DB
func New(rep *sqlx.DB, exchangeFolder string) dto.Repository {
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
