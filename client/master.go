package client

import (
	"context"
	"fmt"
	"sync"

	"github.com/egorka-gh/zbazar/zsync/client/http"
	http1 "github.com/egorka-gh/zbazar/zsync/pkg/http"
	"github.com/egorka-gh/zbazar/zsync/pkg/service"
)

//sync Master version
func (c *Client) syncMaster(ctx context.Context) (e1 error) {
	defer func() {
		c.logger.Log("method", "Sync", "e1", e1)
	}()
	if ctx.Err() != nil {
		//context canceled
		return ctx.Err()
	}

	src, e1 := c.db.ListSource(ctx, c.id)
	if e1 != nil {
		return e1
	}

	wg := sync.WaitGroup{}

	//TODO get versioned tables count from db
	const tablesNum int = 1
	//pulled packs chan
	pulled := make(chan pack, tablesNum*len(src))
	//downloaded packs chan
	loaded := make(chan pack, tablesNum)

	//start database worker
	wg.Add(1)
	go func() {
		for p := range loaded {
			if p.Err != nil {
				c.logger.Log("Sync", "load", "url", p.URL+http1.PackPattern+p.Pack.Pack, "size_kb", fmt.Sprintf("%.2f", float32(p.Pack.PackSize)/1024), "e1", p.Err)
			} else {
				//exec in db
				p.Err = c.db.ExecPack(ctx, p.Pack)
				c.logger.Log("Sync", "exec", "pack", p.Pack.Pack, "size_kb", fmt.Sprintf("%.2f", float32(p.Pack.PackSize)/1024), "e1", p.Err)
			}
			if p.Svc != nil {
				//notify server can remove pack
				//don't care rusult
				_ = p.Svc.PackDone(ctx, p.Pack)
			}
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

	//start pull workers
	wg.Add(1)
	//pull version packs from each source
	wgs := sync.WaitGroup{}
	//TODO limit workers?
	for _, s := range src {
		//start pull worker for source
		c.logger.Log("Sync", "start_pull", "source", s.ID, "url", s.URL)
		svc, err := http.New(s.URL, defaultHttpOptions(c.logger))
		if e1 != nil {
			c.logger.Log("Sync", "start_pull", "source", s.ID, "url", s.URL, "e1", err)
		} else {
			wgs.Add(1)
			go func(s service.Source, svc service.ZsyncService) {
				defer wgs.Done()
				_ = c.pullSyncPacks(ctx, svc, s.ID, s.URL, pulled)
			}(s, svc)
		}
	}

	//waite pull workers
	go func() {
		wgs.Wait()
		close(pulled)
		wg.Done()
	}()

	//waite all downloads complete & close loaded chan
	go func() {
		wgl.Wait()
		close(loaded)
		wg.Done()
	}()

	//waite database worker
	wg.Wait()
	return e1
}
