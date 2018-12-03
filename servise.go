package zsync

import (
	"context"
	"errors"

	"github.com/egorka-gh/zbazar/zsync/dto"
)

var (
	ErrOrderNotFound   = errors.New("not found")
	ErrCmdRepository   = errors.New("unable to command repository")
	ErrQueryRepository = errors.New("unable to query repository")
)

// Service describes the zsync service.
type Service interface {
	ListVersion(ctx context.Context, source string) ([]dto.Version, error)
	GetPack(ctx context.Context, source string, table string, start int) (dto.VersionPack, error)
	PushPack(ctx context.Context, pack dto.VersionPack) error
	//cash
	AddActivity(ctx context.Context, activity dto.Activity) error
	GetLevel(ctx context.Context, card string) (int, error)
}
