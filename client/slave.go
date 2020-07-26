package client

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/egorka-gh/zbazar/zsync/client/http"
	http1 "github.com/egorka-gh/zbazar/zsync/pkg/http"
)

//sync Subordinate version
func (c *Client) syncSubordinate(ctx context.Context) (e1 error) {
	defer func() {
		c.logger.Log("method", "Sync", "e1", e1)
	}()
	if ctx.Err() != nil {
		//context canceled
		return ctx.Err()
	}
	c.logger.Log("Sync", "start_pull", "source", "00", "url", c.mainURL)

	if c.mainURL == "" {
		e1 = errors.New("Empty main URL")
		return e1
	}

	svc, e1 := http.New(c.mainURL, defaultHttpOptions(c.logger))
	if e1 != nil {
		return e1
	}

	//TODO get versioned tables count from db
	const tablesNum int = 10
	wg := sync.WaitGroup{}

	//pulled packs chan
	pulled := make(chan pack, tablesNum)
	//downloaded packs chan
	loaded := make(chan pack, tablesNum)

	//start database worker
	wg.Add(1)
	go func() {
		for p := range loaded {
			if p.Err != nil {
				c.logger.Log("Sync", "load", "url", p.URL+http1.PackPattern+p.Pack.Pack, "size_kb", fmt.Sprintf("%.2f", float32(p.Pack.PackSize)/1024), "e1", p.Err)
			} else {
				p.Err = c.db.ExecPack(ctx, p.Pack)
				c.logger.Log("Sync", "exec", "pack", p.Pack.Pack, "size_kb", fmt.Sprintf("%.2f", float32(p.Pack.PackSize)/1024), "e1", p.Err)
			}
			//notify server can remove pack
			//don't care rusult
			_ = svc.PackDone(ctx, p.Pack)
		}
		wg.Done()
	}()

	//start 5 loaders
	wg.Add(1)
	wgl := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wgl.Add(1)
		go func() {
			c.syncPackloader(ctx, pulled, loaded)
			wgl.Done()
		}()
	}

	//start pull worker
	wg.Add(1)
	go func() {
		defer func() {
			close(pulled)
			wg.Done()
		}()
		_ = c.pullSyncPacks(ctx, svc, "00", c.mainURL, pulled)
	}()

	//waite all downloads complete & close loaded chan
	go func() {
		wgl.Wait()
		close(loaded)
		wg.Done()
	}()

	wg.Wait()
	return e1
}
