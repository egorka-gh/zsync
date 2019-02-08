package scheduler

import (
	"context"
	"errors"
	"time"
)

//Scheduler runs tasks at specified date/time or at specified interval
type Scheduler struct {
	tasks  []*task
	cancel context.CancelFunc
}

type task struct {
	scheduleType int
	execute      func(ctx context.Context) error
	interval     time.Duration
	day          int
	hour         int
	lastRun      time.Time
}

const typePeriodic int = 1
const typeDaily int = 2
const typeMonthly int = 3

const ticInterval time.Duration = time.Minute

//AddPeriodic adds task that runs at specified interval
func (s *Scheduler) AddPeriodic(interval time.Duration, execute func(ctx context.Context) error) {
	s.tasks = append(s.tasks,
		&task{
			scheduleType: typePeriodic,
			execute:      execute,
			interval:     interval,
		})
}

//AddDaily adds task that runs daily at specified hour
func (s *Scheduler) AddDaily(hour int, execute func(ctx context.Context) error) {
	s.tasks = append(s.tasks,
		&task{
			scheduleType: typeDaily,
			execute:      execute,
			hour:         hour,
		})
}

//AddMonthly adds task that runs Monthly at specified day, hour
func (s *Scheduler) AddMonthly(day, hour int, execute func(ctx context.Context) error) {
	s.tasks = append(s.tasks,
		&task{
			scheduleType: typeMonthly,
			execute:      execute,
			day:          day,
			hour:         hour,
		})
}

//Run starts Scheduler
//every ticInterval it checks/run tasks in series
func (s *Scheduler) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	if len(s.tasks) == 0 {
		return errors.New("Empty Scheduler task list")
	}

	for alive := true; alive; {
		for _, t := range s.tasks {
			//ckeck task
			switch t.scheduleType {
			case typePeriodic:
				if t.lastRun.IsZero() || time.Since(t.lastRun) >= t.interval {
					t.execute(ctx)
					t.lastRun = time.Now()
				}
			case typeDaily:
				now := time.Now()
				year, month, day := now.Date()
				midnight := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
				if now.Hour() >= t.hour && (t.lastRun.IsZero() || t.lastRun.Before(midnight)) {
					t.execute(ctx)
					t.lastRun = time.Now()
				}
			case typeMonthly:
				now := time.Now()
				year, month, day := now.Date()
				midnight := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
				if day == t.day && now.Hour() >= t.hour && (t.lastRun.IsZero() || t.lastRun.Before(midnight)) {
					t.execute(ctx)
					t.lastRun = time.Now()
				}
			}
		}
		//check canceled
		if ctx.Err() != nil && (ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded) {
			alive = false
		} else {
			//waite tick
			timer := time.NewTimer(ticInterval)
			select {
			case <-ctx.Done():
				timer.Stop()
				alive = false
			case <-timer.C:
				//new cycle
			}
		}
	}
	return nil
}

//Stop stops Scheduler, blocking while started task complite
func (s *Scheduler) Stop() {
	s.cancel()
}
