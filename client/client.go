package client

import (
	"context"
	"crypto/md5"
	"encoding/hex"

	"github.com/cavaliercoder/grab"
	"github.com/go-kit/kit/log"

	"github.com/egorka-gh/zbazar/zsync/pkg/service"
)

//Client works vs ZsyncService
type Client struct {
	db        service.Repository
	id        string
	packURL   string
	masterURL string
	logger    log.Logger
}

//NewMaster creates client
func NewMaster(rep service.Repository, id, packURL string, logger log.Logger) *Client {
	return &Client{
		db:      rep,
		id:      id,
		packURL: packURL,
		logger:  logger,
	}
}

//NewSlave creates client
func NewSlave(rep service.Repository, id, masterURL, packURL string, logger log.Logger) *Client {
	return &Client{
		db:        rep,
		id:        id,
		packURL:   packURL,
		logger:    logger,
		masterURL: masterURL,
	}
}

//internal exchange struct
type pack struct {
	Pack service.VersionPack
	URL  string
	Err  error
	Svc  service.ZsyncService
}

//Sync perfoms synchronization
func (c *Client) Sync(ctx context.Context) (e1 error) {
	if c.id == "00" {
		return c.syncMaster(ctx)
	}
	return c.syncSlave(ctx)
}

//TODO recheck source usage
//PullSyncPacks checks remote versions and download version pakcs
func (c *Client) pullSyncPacks(ctx context.Context, svc service.ZsyncService, source string, url string, out chan<- pack) (e1 error) {
	defer func() {
		c.logger.Log("thread", "client", "method", "PullSyncPacks", "source", source, "url", url, "e1", e1)
	}()

	//get remote versions
	vr, e1 := svc.ListVersion(ctx, c.id)
	if e1 != nil {
		return
	}

	//load lockal versions
	vl, e1 := c.db.ListVersion(ctx, source)
	if e1 != nil {
		return
	}

	//compare versions
	for _, v0 := range vr {
		for _, v1 := range vl {
			if v0.Source == v1.Source && v0.Table == v1.Table && v0.Version > v1.Version {
				ch := make(chan pack, 1)
				//serial
				//TODO parallel goroutines?
				//pull pack from remote
				go func(v service.Version) {
					vp, err := svc.PullPack(ctx, v.Source, v.Table, v.Version)
					ch <- pack{Pack: vp, URL: url, Err: err, Svc: svc}
				}(v1)
				select {
				case <-ctx.Done():
					<-ch // Wait for client
					//fmt.Println("Cancel the context")
					e1 = ctx.Err()
					return
				case data := <-ch:
					if data.Err != nil {
						c.logger.Log("thread", "client", "method", "PullSyncPacks", "source", data.Pack.Source, "url", url, "table", data.Pack.Table, "e1", data.Err)
						if e1 == nil {
							e1 = data.Err
						}
					} else {
						//check if has sync data
						if data.Pack.Pack != "" {
							out <- data
						}
					}
				}
			}
		}
	}
	return
}

//loadSyncPack pack download worker
func (c *Client) loadSyncPack(ctx context.Context, client *grab.Client, in <-chan pack, out chan<- pack) {
	select {
	case <-ctx.Done():
		//context canceled
		return
	case p, ok := <-in:
		if !ok {
			//chan closed
			return
		}
		//create request
		req, err := grab.NewRequest(c.db.ExchangeFolder(), p.URL+c.packURL+"/"+p.Pack.Pack)
		if err != nil {
			p.Err = err
		} else {
			req.Size = p.Pack.PackSize
			b, err := hex.DecodeString(p.Pack.PackMD5)
			if err != nil {
				p.Err = err
			} else {
				req.SetChecksum(md5.New(), b, true)
				//cancelabel
				req = req.WithContext(ctx)
				//load
				resp := client.Do(req)
				//respch <- resp
				//waite complite
				//<-resp.Done
				p.Err = resp.Err()
				//check if context canceled
				if ctx.Err() != nil && (ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded) {
					return
				}
			}
		}
		//sent result
		out <- p
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
