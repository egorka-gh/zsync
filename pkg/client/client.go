package client

import (
	"context"

	"github.com/egorka-gh/zbazar/zsync/pkg/service"
)

//Client works vs ZsyncService
type Client struct {
	db       service.Repository
	id       string
	endpoint string
}

//New creates client
func New(rep service.Repository, id, endpoint string) *Client {
	return &Client{
		db:       rep,
		id:       id,
		endpoint: endpoint,
	}
}

//FixVersions updates versions in db
func (c *Client) FixVersions(ctx context.Context) (e0 error) {
	return c.db.FixVersions(ctx, c.id)
}

//ExecPack apply version pack in db
func (c *Client) ExecPack(ctx context.Context, pack service.VersionPack) (e0 error) {
	return c.db.ExecPack(ctx, pack)
}
