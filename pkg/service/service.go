package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"strconv"
)

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
	db             Repository
	id             string
	exchangeFolder string
}

func (b *basicZsyncService) fileInfo(fileName string) (size int64, md5Str string, err error) {
	var path = b.exchangeFolder + fileName
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return 0, "", err
	}
	fi, err := file.Stat()
	if err != nil {
		return 0, "", err
	}
	size = fi.Size()

	if size > 0 {
		hash := md5.New()
		_, err = io.Copy(hash, file)
		if err != nil {
			return 0, "", err
		}
		hashBytes := hash.Sum(nil)[:16]
		md5Str = hex.EncodeToString(hashBytes)
	}

	return size, md5Str, nil
}

func (b *basicZsyncService) ListVersion(ctx context.Context, source string) (v0 []Version, e1 error) {
	return b.db.ListVersion(ctx, b.id)
}

func (b *basicZsyncService) PullPack(ctx context.Context, source string, table string, start int) (v0 VersionPack, e1 error) {
	//pack name vs asker (source) prifix and start version sifix
	var fileName = source + "_" + table + "_" + strconv.FormatInt(int64(start), 10) + ".dat"
	v0, e1 = b.db.CreatePack(ctx, b.id, table, fileName, start)
	if e1 != nil {
		return v0, e1
	}
	//get check filesize, calc MD5
	if v0.Pack != "" {
		v0.PackSize, v0.PackMD5, e1 = b.fileInfo(v0.Pack)
		if e1 != nil {
			return v0, e1
		}
		if v0.PackSize == 0 {
			//empty version (no changes, but version fixed), posible bug
			_ = b.db.DelPack(ctx, v0)
			v0.End = v0.Start
			v0.Pack = ""
		}
	}

	return v0, e1
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
func NewBasicZsyncService(rep Repository, id, folder string) ZsyncService {
	return &basicZsyncService{
		db:             rep,
		id:             id,
		exchangeFolder: folder,
	}
}

// New returns a ZsyncService with all of the expected middleware wired in.
func New(middleware []Middleware, rep Repository, id string, folder string) ZsyncService {
	var svc ZsyncService = NewBasicZsyncService(rep, id, folder)
	for _, m := range middleware {
		svc = m(svc)
	}
	return svc
}
