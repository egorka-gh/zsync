package scheduler

import (
	"context"
	"errors"
	"time"
)

//Scheduler runs tasks at specified date/time or at specified interval (unusable after stop)
type Scheduler interface {
	AddPeriodic(interval time.Duration, execute func(ctx context.Context) error)
	AddDaily(hour int, execute func(ctx context.Context) error)
	AddMonthly(day, hour int, execute func(ctx context.Context) error)
	Run() error
	Stop()
}

type basicScheduler struct {
	tasks  []*task
	stopCh chan struct{}
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

//New creates new  Scheduler
func New() Scheduler {
	stopCh := make(chan struct{})
	return &basicScheduler{
		stopCh: stopCh,
	}

}

//AddPeriodic adds task that runs at specified interval
func (s *basicScheduler) AddPeriodic(interval time.Duration, execute func(ctx context.Context) error) {
	s.tasks = append(s.tasks,
		&task{
			scheduleType: typePeriodic,
			execute:      execute,
			interval:     interval,
		})
}

//AddDaily adds task that runs daily at specified hour
func (s *basicScheduler) AddDaily(hour int, execute func(ctx context.Context) error) {
	s.tasks = append(s.tasks,
		&task{
			scheduleType: typeDaily,
			execute:      execute,
			hour:         hour,
		})
}

//AddMonthly adds task that runs Monthly at specified day, hour
func (s *basicScheduler) AddMonthly(day, hour int, execute func(ctx context.Context) error) {
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
func (s *basicScheduler) Run() error {
	//TODO check if starting stopped Scheduler
	if len(s.tasks) == 0 {
		return errors.New("Empty Scheduler task list")
	}

	for alive := true; alive; {
		for _, t := range s.tasks {
			runIt := false
			//ckeck task
			switch t.scheduleType {
			case typePeriodic:
				runIt = t.lastRun.IsZero() || time.Since(t.lastRun) >= t.interval
			case typeDaily:
				now := time.Now()
				year, month, day := now.Date()
				midnight := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
				runIt = now.Hour() >= t.hour && (t.lastRun.IsZero() || t.lastRun.Before(midnight))
			case typeMonthly:
				now := time.Now()
				year, month, day := now.Date()
				midnight := time.Date(year, month, day, 0, 0, 0, 0, now.Location())
				runIt = day == t.day && now.Hour() >= t.hour && (t.lastRun.IsZero() || t.lastRun.Before(midnight))
			}
			if runIt {
				//run task
				//create task context
				ctx, cancel := context.WithCancel(context.Background())
				//monitor stop chan
				complite := make(chan struct{})
				go func(cancel context.CancelFunc) {
					select {
					case <-s.stopCh:
						cancel()
						//race ok
						alive = false
					case <-complite:
					}
				}(cancel)
				//run task
				t.execute(ctx)
				t.lastRun = time.Now()
				close(complite)
			}
			if !alive {
				break
			}
		}

		if alive {
			//waite tick or stop
			timer := time.NewTimer(ticInterval)
			select {
			case <-s.stopCh:
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
func (s *basicScheduler) Stop() {
	//s.cancel()
	close(s.stopCh)
}
