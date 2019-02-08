package scheduler

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

//go test -timeout 10m github.com\egorka-gh\zbazar\zsync\pkg\scheduler -run ^(TestShedule)$
func TestShedule(t *testing.T) {
	pCount, dCount, mCount, mwCount := 0, 0, 0, 0

	sh := Scheduler{}
	sh.AddPeriodic(3*time.Minute,
		func(ctx context.Context) error {
			pCount++
			fmt.Println("Doing Periodic")
			t := time.NewTimer(2 * time.Second)
			select {
			case <-ctx.Done():
				fmt.Println("Periodic canceled")
				t.Stop()
			case <-t.C:
				fmt.Println("Complite Periodic")
			}
			return nil
		},
	)

	sh.AddDaily(time.Now().Hour(),
		func(ctx context.Context) error {
			dCount++
			fmt.Println("Doing Daily")
			t := time.NewTimer(2 * time.Second)
			select {
			case <-ctx.Done():
				fmt.Println("Daily canceled")
				t.Stop()
			case <-t.C:
				fmt.Println("Complite Daily")
			}
			return nil
		},
	)

	sh.AddMonthly(time.Now().Day(), time.Now().Hour(),
		func(ctx context.Context) error {
			mCount++
			fmt.Println("Doing Monthly")
			t := time.NewTimer(2 * time.Second)
			select {
			case <-ctx.Done():
				fmt.Println("Monthly canceled")
				t.Stop()
			case <-t.C:
				fmt.Println("Complite Monthly")
			}
			return nil
		},
	)

	sh.AddMonthly(time.Now().Day()+1, time.Now().Hour(),
		func(ctx context.Context) error {
			mwCount++
			fmt.Println("Doing Wrong Monthly")
			t := time.NewTimer(10 * time.Second)
			select {
			case <-ctx.Done():
				fmt.Println("Wrong Monthly canceled")
				t.Stop()
			case <-t.C:
				fmt.Println("Complite Wrong Monthly")
			}
			return nil
		},
	)

	w := sync.WaitGroup{}
	w.Add(1)
	go func() {
		sh.Run()
		w.Done()
	}()

	time.AfterFunc(7*time.Minute, sh.Stop)
	w.Wait()
	if pCount != 3 {
		t.Error("Periodic wrong executes count, expected 3, got", pCount)
	}
	if dCount != 1 {
		t.Error("Daily wrong executes count, expected 1, got", dCount)
	}
	if mCount != 1 {
		t.Error("Monthly wrong executes count, expected 1, got", mCount)
	}
	if mwCount != 0 {
		t.Error("Monthly (tommorow) wrong executes count, expected 0, got", mwCount)
	}
}
