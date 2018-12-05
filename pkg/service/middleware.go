package service

import (
	"context"
	log "github.com/go-kit/kit/log"
)

// Middleware describes a service middleware.
type Middleware func(ZsyncService) ZsyncService

type loggingMiddleware struct {
	logger log.Logger
	next   ZsyncService
}

// LoggingMiddleware takes a logger as a dependency
// and returns a ZsyncService Middleware.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next ZsyncService) ZsyncService {
		return &loggingMiddleware{logger, next}
	}

}

func (l loggingMiddleware) ListVersion(ctx context.Context, source string) (v0 []Version, e1 error) {
	defer func() {
		l.logger.Log("method", "ListVersion", "source", source, "v0", v0, "e1", e1)
	}()
	return l.next.ListVersion(ctx, source)
}
func (l loggingMiddleware) PullPack(ctx context.Context, source string, table string, start int) (v0 VersionPack, e1 error) {
	defer func() {
		l.logger.Log("method", "PullPack", "source", source, "table", table, "start", start, "v0", v0, "e1", e1)
	}()
	return l.next.PullPack(ctx, source, table, start)
}
func (l loggingMiddleware) PushPack(ctx context.Context, pack VersionPack) (e0 error) {
	defer func() {
		l.logger.Log("method", "PushPack", "pack", pack, "e0", e0)
	}()
	return l.next.PushPack(ctx, pack)
}
func (l loggingMiddleware) PackDone(ctx context.Context, pack VersionPack) (e0 error) {
	defer func() {
		l.logger.Log("method", "PackDone", "pack", pack, "e0", e0)
	}()
	return l.next.PackDone(ctx, pack)
}
func (l loggingMiddleware) AddActivity(ctx context.Context, activity Activity) (e0 error) {
	defer func() {
		l.logger.Log("method", "AddActivity", "activity", activity, "e0", e0)
	}()
	return l.next.AddActivity(ctx, activity)
}
func (l loggingMiddleware) GetLevel(ctx context.Context, card string) (i0 int, e1 error) {
	defer func() {
		l.logger.Log("method", "GetLevel", "card", card, "i0", i0, "e1", e1)
	}()
	return l.next.GetLevel(ctx, card)
}
