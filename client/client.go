package client

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"time"

	"github.com/cavaliercoder/grab"
	"github.com/go-kit/kit/log"

	"github.com/egorka-gh/zbazar/zsync/pkg/http"
	"github.com/egorka-gh/zbazar/zsync/pkg/service"
	http1 "github.com/go-kit/kit/transport/http"
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

	if ctx.Err() != nil {
		return ctx.Err()
	}
	//get remote versions
	vr, e1 := svc.ListVersion(ctx, c.id)
	if e1 != nil {
		return e1
	}
	if len(vr) == 0 {
		//empty response or transport error
		return errors.New("Empty response")
	}
	//load lockal versions
	vl, e1 := c.db.ListVersion(ctx, source)
	if e1 != nil {
		return e1
	}

	//compare versions and do requests in serial
	for _, v0 := range vr {
		for _, v1 := range vl {
			if v0.Source == v1.Source && v0.Table == v1.Table && v0.Version > v1.Version {
				//check context canceled
				if ctx.Err() != nil {
					return ctx.Err()
				}
				c.logger.Log("method", "PullSyncPacks", "source", source, "table", v0.Table, "remote_version", v0.Version, "lockal_version", v1.Version)
				//pull pack from remote
				vp, err := svc.PullPack(ctx, c.id, v1.Table, v1.Version)
				c.logger.Log("method", "PullSyncPacks", "source", source, "table", v0.Table, "pack", vp.Pack, "start", vp.Start, "end", vp.End, "err", err)
				//check if no errors and has sync data
				if err == nil {
					if vp.Pack != "" {
						out <- pack{Pack: vp, URL: url, Err: err, Svc: svc}
					}
				} else {
					//fix first err
					if e1 == nil {
						e1 = err
					}
				}
			}
		}
	}
	return e1
}

//loadSyncPack pack download worker
func (c *Client) syncPackloader(ctx context.Context, client *grab.Client, in <-chan pack, out chan<- pack) {
	for p := range in {
		//check if contex canceled
		if ctx.Err() != nil {
			p.Err = ctx.Err()
			out <- p
			continue
		}
		//create request
		req, err := grab.NewRequest(c.db.ExchangeFolder(), p.URL+http.PackPattern+p.Pack.Pack)
		if err != nil {
			p.Err = err
			out <- p
			continue
		}
		req.Size = p.Pack.PackSize
		b, err := hex.DecodeString(p.Pack.PackMD5)
		if err != nil {
			p.Err = err
			out <- p
			continue
		}
		req.SetChecksum(md5.New(), b, true)
		//cancelabel
		req = req.WithContext(ctx)
		//load
		resp := client.Do(req)
		//respch <- resp
		//waite complite
		<-resp.Done
		p.Err = resp.Err()
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

func defaultHttpOptions(logger log.Logger) map[string][]http1.ClientOption {
	options := map[string][]http1.ClientOption{
		"AddActivity": {clientFinalizer("AddActivity", logger)},
		"GetLevel":    {clientFinalizer("GetLevel", logger)},
		"ListVersion": {clientFinalizer("ListVersion", logger)},
		"PackDone":    {clientFinalizer("PackDone", logger)},
		"PullPack":    {clientFinalizer("PullPack", logger)},
		"PushPack":    {clientFinalizer("PushPack", logger)},
	}
	return options
}

func clientFinalizer(method string, logger log.Logger) http1.ClientOption {
	lg := log.With(logger, "method", method)
	return http1.ClientFinalizer(
		func(ctx context.Context, err error) {
			if err != nil {
				lg.Log("transport_error", err)
			}
		},
	)
}
