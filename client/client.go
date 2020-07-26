package client

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	http0 "net/http"
	"os"
	"time"

	"github.com/hashicorp/go-cleanhttp"

	"github.com/go-kit/kit/log"

	"github.com/egorka-gh/zbazar/zsync/pkg/http"
	"github.com/egorka-gh/zbazar/zsync/pkg/service"
	http1 "github.com/go-kit/kit/transport/http"
)

//Client works vs ZsyncService
type Client struct {
	db        service.Repository
	id        string
	mainURL string
	logger    log.Logger
}

//NewMain creates client
func NewMain(rep service.Repository, id string, logger log.Logger) *Client {
	cliLog := log.With(logger, "thread", "client")
	return &Client{
		db:     rep,
		id:     id,
		logger: cliLog,
	}
}

//NewSubordinate creates client
func NewSubordinate(rep service.Repository, id, mainURL string, logger log.Logger) *Client {
	cliLog := log.With(logger, "thread", "client")
	return &Client{
		db:        rep,
		id:        id,
		logger:    cliLog,
		mainURL: mainURL,
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
		return c.syncMain(ctx)
	}
	return c.syncSubordinate(ctx)
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
func (c *Client) syncPackloader(ctx context.Context, in <-chan pack, out chan<- pack) {
	for p := range in {
		//check if contex canceled
		if ctx.Err() != nil {
			p.Err = ctx.Err()
			out <- p
			continue
		}

		p.Err = c.downloadPack(ctx, p.URL, p.Pack.Pack)
		if p.Err != nil {
			out <- p
			continue
		}
		//TODO check ctx canceled?
		p.Err = c.checkPack(p.Pack.PackSize, p.Pack.PackMD5, p.Pack.Pack)
		out <- p
	}
}

//TODO implement as service method
func (c *Client) downloadPack(ctx context.Context, baseURL, fileName string) error {

	if fileName == "" {
		return errors.New("empty file name")
	}

	// Create the file
	path := c.db.ExchangeFolder() + fileName
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	url := baseURL + http.PackPattern + fileName
	cli := defaultHttpClient()

	req, err := http0.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		resp.Body.Close()
		return fmt.Errorf("bad response code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

//TODO context cancel??
func (c *Client) checkPack(fileSize int64, fileHash, fileName string) error {
	//check size && hash
	path := c.db.ExchangeFolder() + fileName
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		return err
	}
	if fileSize != fi.Size() {
		return errors.New("file size mismatch")
	}

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return err
	}
	hashBytes := hash.Sum(nil)[:16]
	//TODO do it before download
	hashExpect, err := hex.DecodeString(fileHash)
	if err != nil {
		return err
	}
	if !bytes.Equal(hashExpect, hashBytes) {
		return errors.New("checksum mismatch")
	}
	return nil
}

//FixVersions updates versions in db
func (c *Client) FixVersions(ctx context.Context) (e0 error) {
	defer func() {
		c.logger.Log("method", "FixVersions", "error", e0)
	}()
	return c.db.FixVersions(ctx, c.id)
}

//CleanUp - delete not relevant data (old balance etc)
func (c *Client) CleanUp(ctx context.Context) (e0 error) {
	defer func() {
		c.logger.Log("method", "CleanUp", "error", e0)
	}()
	return c.db.CleanUp(ctx)
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

//creates transient client, can be pooled?
func defaultHttpClient() *http0.Client {
	cli := cleanhttp.DefaultClient()
	cli.Timeout = time.Minute * 3
	return cli
}

func defaultHttpOptions(logger log.Logger) map[string][]http1.ClientOption {
	cli := defaultHttpClient()
	options := map[string][]http1.ClientOption{
		"AddActivity": {http1.SetClient(cli), clientFinalizer("AddActivity", logger)},
		"GetLevel":    {http1.SetClient(cli), clientFinalizer("GetLevel", logger)},
		"ListVersion": {http1.SetClient(cli), clientFinalizer("ListVersion", logger)},
		"PackDone":    {http1.SetClient(cli), clientFinalizer("PackDone", logger)},
		"PullPack":    {http1.SetClient(cli), clientFinalizer("PullPack", logger)},
		"PushPack":    {http1.SetClient(cli), clientFinalizer("PushPack", logger)},
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
