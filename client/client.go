package client

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/cavaliercoder/grab"
	"github.com/go-kit/kit/log"

	"github.com/egorka-gh/zbazar/zsync/client/http"
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

//New creates client
func New(rep service.Repository, id, packURL string, logger log.Logger) *Client {
	return &Client{
		db:      rep,
		id:      id,
		packURL: packURL,
		logger:  logger,
	}
}

//internal exchange struct
type pack struct {
	Pack service.VersionPack
	URL  string
	Err  error
}

//Sync perfoms synchronization
//Slave version
func (c *Client) Sync(ctx context.Context) (e1 error) {
	defer func() {
		c.logger.Log("thread", "client", "method", "Sync", "e1", e1)
	}()
	c.logger.Log("thread", "client", "method", "Sync", "operation", "start", "source", "00", "url", c.masterURL)

	if c.masterURL == "" {
		e1 = errors.New("Empty master URL")
		return
	}

	svc, e1 := http.New(c.masterURL, nil)
	if e1 != nil {
		return
	}

	//TODO get versioned tables count from db
	const tablesNum int = 10
	wg := sync.WaitGroup{}

	//pull version packs
	pulled := make(chan pack, tablesNum)
	//pullwg := sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() {
			close(pulled)
			wg.Done()
		}()
		_ = c.pullSyncPacks(ctx, svc, "00", c.masterURL, pulled)
	}()

	//download packs
	loaded := make(chan pack, tablesNum)

	client := grab.NewClient()
	load := func(in <-chan pack, out chan<- pack) {
		// TODO: enable cancelling of batch jobs
		for p := range in {
			//create request
			req, err := grab.NewRequest(c.db.ExchangeFolder(), c.masterURL+c.packURL+"/"+p.Pack.Pack)
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
					out <- p
				}
			}
		}
	}
	// start 5 loaders
	wgl := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wgl.Add(1)
		go func() {
			load(pulled, loaded)
			wgl.Done()
		}()
	}

	//waite all downloads complete & close loaded chan
	go func() {
		wgl.Wait()
		close(loaded)
	}()

	//exec in db
	wg.Add(1)
	go func() {
		for p := range loaded {
			if p.Err != nil {
				c.logger.Log("thread", "client", "method", "Sync", "operation", "load", "url", c.masterURL+c.packURL+"/"+p.Pack.Pack, "e1", p.Err)
			} else {
				p.Err = c.db.ExecPack(ctx, p.Pack)
				if p.Err != nil {
					c.logger.Log("thread", "client", "method", "Sync", "operation", "exec", "pack", p.Pack.Pack, "e1", p.Err)
				}
			}
			//notify server can remove pack
			//don't care rusult
			go func(p pack) {
				_ = svc.PackDone(ctx, p.Pack)
			}(p)
		}
		wg.Done()
	}()

	wg.Wait()
	return
}

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
					ch <- pack{Pack: vp, URL: url, Err: err}
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

//FixVersions updates versions in db
func (c *Client) FixVersions(ctx context.Context) (e0 error) {
	return c.db.FixVersions(ctx, c.id)
}

//ExecPack apply version pack in db
func (c *Client) ExecPack(ctx context.Context, pack service.VersionPack) (e0 error) {
	return c.db.ExecPack(ctx, pack)
}
