package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(16*time.Second, cancel)
	for alive := true; alive; {

		func(ctx context.Context) {
			fmt.Println("Doing some work")
			t := time.NewTimer(2 * time.Second)
			select {
			case <-ctx.Done():
				fmt.Println("Work canceled")
				t.Stop()
			case <-t.C:
				fmt.Println("Complite work")
			}
		}(ctx)

		if ctx.Err() != nil && (ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded) {
			fmt.Printf("Stopping (context canceled): %v\n", ctx.Err())
			alive = false
		} else {
			timer := time.NewTimer(5 * time.Second)
			fmt.Println("sleeping")
			select {
			case <-ctx.Done():
				timer.Stop()
				alive = false
				fmt.Println("Canceled while sleeping")
			case <-timer.C:
				//rerun work
			}

		}
	}

	fs := http.FileServer(http.Dir("D:\\Buffer\\zexch\\zs\\log\\"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Listening...")
	http.ListenAndServe(":3000", nil)
}
