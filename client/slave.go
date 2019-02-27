package client

import (
	"context"
	"errors"
	"sync"

	"github.com/egorka-gh/zbazar/zsync/client/http"
	http1 "github.com/egorka-gh/zbazar/zsync/pkg/http"
)

//sync Slave version
func (c *Client) syncSlave(ctx context.Context) (e1 error) {
	defer func() {
		c.logger.Log("method", "Sync", "e1", e1)
	}()
	if ctx.Err() != nil {
		//context canceled
		return ctx.Err()
	}
	c.logger.Log("method", "Sync", "operation", "start", "source", "00", "url", c.masterURL)

	if c.masterURL == "" {
		e1 = errors.New("Empty master URL")
		return e1
	}

	svc, e1 := http.New(c.masterURL, defaultHttpOptions(c.logger))
	if e1 != nil {
		return e1
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

	// start 5 loaders
	wgl := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wgl.Add(1)
		go func() {
			//load(pulled, loaded)
			c.syncPackloader(ctx, pulled, loaded)
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
				c.logger.Log("method", "Sync", "operation", "load", "url", p.URL+http1.PackPattern+p.Pack.Pack, "e1", p.Err)
			} else {
				p.Err = c.db.ExecPack(ctx, p.Pack)
				c.logger.Log("method", "Sync", "operation", "exec", "pack", p.Pack.Pack, "e1", p.Err)
			}
			//notify server can remove pack
			//don't care rusult
			_ = svc.PackDone(ctx, p.Pack)
		}
		wg.Done()
	}()

	wg.Wait()
	return e1
}
