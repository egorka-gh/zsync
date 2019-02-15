package client

import (
	"context"
	"sync"

	"github.com/cavaliercoder/grab"

	"github.com/egorka-gh/zbazar/zsync/client/http"
	http1 "github.com/egorka-gh/zbazar/zsync/pkg/http"
	"github.com/egorka-gh/zbazar/zsync/pkg/service"
)

//sync Master version
func (c *Client) syncMaster(ctx context.Context) (e1 error) {
	defer func() {
		c.logger.Log("method", "Sync", "e1", e1)
	}()

	//TODO get versioned tables count from db
	const tablesNum int = 1
	src, e1 := c.db.ListSource(ctx, c.id)
	if e1 != nil {
		return
	}

	wg := sync.WaitGroup{}

	//pull version packs from each source
	pulled := make(chan pack, tablesNum*len(src))
	wg.Add(1)

	wgs := sync.WaitGroup{}
	//TODO limit workers?
	for _, s := range src {
		//start pull worker for source
		c.logger.Log("method", "Sync", "operation", "start", "source", s.ID, "url", s.URL)
		svc, err := http.New(s.URL, nil)
		if e1 != nil {
			c.logger.Log("method", "Sync", "operation", "start", "source", s.ID, "url", s.URL, "e1", err)
		} else {
			wgs.Add(1)
			go func(url string, svc service.ZsyncService) {
				defer wgs.Done()
				_ = c.pullSyncPacks(ctx, svc, c.id, url, pulled)
			}(s.URL, svc)
		}
	}

	//waite pull workers
	go func() {
		wgs.Wait()
		wg.Done()
		close(pulled)
	}()

	//download packs
	loaded := make(chan pack, tablesNum)

	client := grab.NewClient()
	// start 5 loaders
	wgl := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wgl.Add(1)
		go func() {
			c.loadSyncPack(ctx, client, pulled, loaded)
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
				if p.Err != nil {
					c.logger.Log("method", "Sync", "operation", "exec", "pack", p.Pack.Pack, "e1", p.Err)
				}
			}
			if p.Svc != nil {
				//notify server can remove pack
				//don't care rusult
				go func(p pack) {
					_ = p.Svc.PackDone(ctx, p.Pack)
				}(p)
			}
		}
		wg.Done()
	}()

	wg.Wait()
	return
}
