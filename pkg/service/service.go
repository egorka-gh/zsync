package service

import "context"

// ZsyncService describes the service.
type ZsyncService interface {
	ListVersion(ctx context.Context, source string) ([]Version, error)
	PullPack(ctx context.Context, source string, table string, start int) (VersionPack, error)
	PushPack(ctx context.Context, pack VersionPack) error
	PackDone(ctx context.Context, pack VersionPack) error
	//cash
	AddActivity(ctx context.Context, activity Activity) error
	GetLevel(ctx context.Context, card string) (int, error)
}

type basicZsyncService struct {
	db Repository
}

func (b *basicZsyncService) ListVersion(ctx context.Context, source string) (v0 []Version, e1 error) {
	return b.db.ListVersion(ctx, source)
}
func (b *basicZsyncService) PullPack(ctx context.Context, source string, table string, start int) (v0 VersionPack, e1 error) {
	return b.db.CreatePack(ctx, source, table, start)
}
func (b *basicZsyncService) PushPack(ctx context.Context, pack VersionPack) (e0 error) {
	return b.db.ExecPack(ctx, pack)
}
func (b *basicZsyncService) PackDone(ctx context.Context, pack VersionPack) (e0 error) {
	return b.db.DelPack(ctx, pack)
}
func (b *basicZsyncService) AddActivity(ctx context.Context, activity Activity) (e0 error) {
	return b.db.AddActivity(ctx, activity)
}
func (b *basicZsyncService) GetLevel(ctx context.Context, card string) (i0 int, e1 error) {

	return b.db.GetLevel(ctx, card)
}

// NewBasicZsyncService returns a naive, stateless implementation of ZsyncService.
func NewBasicZsyncService(rep Repository) ZsyncService {
	return &basicZsyncService{
		db: rep,
	}
}

// New returns a ZsyncService with all of the expected middleware wired in.
func New(middleware []Middleware, rep Repository) ZsyncService {
	var svc ZsyncService = NewBasicZsyncService(rep)
	for _, m := range middleware {
		svc = m(svc)
	}
	return svc
}
