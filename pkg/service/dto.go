package service

import (
	"context"
	"time"
)

//Version represents the version of db objet
type Version struct {
	Source  string `json:"source"`
	Table   string `json:"table_name"`
	Version int    `json:"latest_version"`
}

//VersionPack represents sync pack
type VersionPack struct {
	Source string `json:"source"`
	Table  string `json:"table_name"`
	Start  int    `json:"start_version"`
	End    int    `json:"latest_version"`
	Pack   string `json:"pack"`
}

//Activity represents card activity
type Activity struct {
	Source    string    `json:"source"`
	Doc       string    `json:"doc_id"`
	Card      string    `json:"card"`
	DocDate   time.Time `json:"doc_date"`
	DocSum    float32   `json:"doc_sum"`
	BonuceSum float32   `json:"bonuce_sum"`
}

// Repository describes the persistence on dto
type Repository interface {
	ListVersion(ctx context.Context, source string) ([]Version, error)
	CreatePack(ctx context.Context, source string, table string, start int) (VersionPack, error)
	ExecPack(ctx context.Context, pack VersionPack) error
	DelPack(ctx context.Context, pack VersionPack) error
	CalcLevels(ctx context.Context, balanceDate time.Time) error
	CalcBalance(ctx context.Context, balanceDate time.Time) error
	FixVersions(ctx context.Context, source string) error
	//cash
	AddActivity(ctx context.Context, activity Activity) error
	GetLevel(ctx context.Context, card string) (int, error)
}
