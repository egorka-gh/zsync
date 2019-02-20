package client

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"time"

	"github.com/cavaliercoder/grab"
	"github.com/go-kit/kit/log"

	"github.com/egorka-gh/zbazar/zsync/pkg/http"
	"github.com/egorka-gh/zbazar/zsync/pkg/service"
)

//Client works vs ZsyncService
type Client struct {
	db        service.Repository
	id        string
	masterURL string
	logger    log.Logger
}

//NewMaster creates client
func NewMaster(rep service.Repository, id string, logger log.Logger) *Client {
	cliLog := log.With(logger, "thread", "client")
	return &Client{
		db:     rep,
		id:     id,
		logger: cliLog,
	}
}

//NewSlave creates client
func NewSlave(rep service.Repository, id, masterURL string, logger log.Logger) *Client {
	cliLog := log.With(logger, "thread", "client")
	return &Client{
		db:        rep,
		id:        id,
		logger:    cliLog,
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
		c.logger.Log("method", "PullSyncPacks", "source", source, "url", url, "e1", e1)
	}()

	//get remote versions
	vr, e1 := svc.ListVersion(ctx, c.id)
	if e1 != nil {
		return e1
	}

	//load lockal versions
	vl, e1 := c.db.ListVersion(ctx, source)
	if e1 != nil {
		return e1
	}

	//compare versions
	for _, v0 := range vr {
		//c.logger.Log("method", "PullSyncPacks", "table", v0.Table, "remote_version", v0.Version)
		for _, v1 := range vl {
			if v0.Source == v1.Source && v0.Table == v1.Table && v0.Version > v1.Version {
				c.logger.Log("method", "PullSyncPacks", "table", v0.Table, "remote_version", v0.Version, "lockal_version", v1.Version)
				ch := make(chan pack, 1)
				//serial
				//TODO parallel goroutines?
				//pull pack from remote
				go func(v service.Version) {
					vp, err := svc.PullPack(ctx, c.id, v.Table, v.Version)
					c.logger.Log("method", "PullSyncPacks", "table", v0.Table, "pack", vp.Pack, "start", vp.Start, "end", vp.End)
					ch <- pack{Pack: vp, URL: url, Err: err, Svc: svc}
				}(v1)
				select {
				case <-ctx.Done():
					<-ch // Wait for client
					//fmt.Println("Cancel the context")
					e1 = ctx.Err()
					return e1
				case data := <-ch:
					if data.Err != nil {
						c.logger.Log("method", "PullSyncPacks", "source", data.Pack.Source, "url", url, "table", data.Pack.Table, "e1", data.Err)
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
	return e1
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
		req, err := grab.NewRequest(c.db.ExchangeFolder(), p.URL+http.PackPattern+p.Pack.Pack)
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
	defer func() {
		c.logger.Log("method", "FixVersions", "error", e0)
	}()
	return c.db.FixVersions(ctx, c.id)
}

//CalcBalance recalcs balance from first day of month
func (c *Client) CalcBalance(ctx context.Context) (e0 error) {
	//get first day of month
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	defer func() {
		c.logger.Log("method", "CalcBalance", "from-date", firstOfMonth, "error", e0)
	}()
	return c.db.CalcBalance(ctx, firstOfMonth)
}

//CalcLevels recalcs levels by previouse month
func (c *Client) CalcLevels(ctx context.Context) (e0 error) {
	//TODO recalc balace??
	//get first day of previouse month
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentMonth = currentMonth - 1
	if currentMonth < 1 {
		currentYear = currentYear - 1
		currentMonth = 12
	}
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	defer func() {
		c.logger.Log("method", "CalcLevels", "from-date", firstOfMonth, "error", e0)
	}()
	return c.db.CalcLevels(ctx, firstOfMonth)
}
