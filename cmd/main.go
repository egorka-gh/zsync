package main

import (
	"errors"
	"log"
	"os"

	"github.com/egorka-gh/zbazar/zsync/cmd/service"
	service2 "github.com/egorka-gh/zbazar/zsync/pkg/service"
	_ "github.com/go-sql-driver/mysql"
	service1 "github.com/kardianos/service"
	group "github.com/oklog/oklog/pkg/group"
)

/*
//Run in terminal
func main() {
	//service.Run()
	service.RunServer()
}
*/

//run as service

var logger service1.Logger

type program struct {
	group     *group.Group
	rep       service2.Repository
	interrupt chan struct{}
	quit      chan struct{}
}

func main() {
	svcConfig := &service1.Config{
		Name:        "ZooSyncServer",
		DisplayName: "Zoo Sync Server",
		Description: "Zoobazar Sync service",
	}
	prg := &program{}

	s, err := service1.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
		return
	}
	if len(os.Args) > 1 {
		err = service1.Control(s, os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}

func (p *program) Start(s service1.Service) error {
	g, rep, err := service.InitServerGroup()
	if err != nil {
		return err
	}

	//defer rep.Close()
	p.group = g
	p.rep = rep
	p.interrupt = make(chan struct{})
	p.quit = make(chan struct{})

	if service1.Interactive() {
		logger.Info("Running in terminal.")
		logger.Infof("Valid startup parametrs: %q\n", service1.ControlAction)
	} else {
		logger.Info("Starting Zsync service...")
	}
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) Stop(s service1.Service) error {
	// Stop should not block. Return with a few seconds.
	logger.Info("Zsync Stopping!")
	//interrupt service
	close(p.interrupt)
	//<-time.After(time.Second * 13)
	//waite service stops
	<-p.quit
	logger.Info("Zsync stopped")
	return nil
}

func (p *program) run() {
	//close db cnn
	defer p.rep.Close()
	running := make(chan struct{})
	//initCancelInterrupt
	p.group.Add(
		func() error {
			select {
			case <-p.interrupt:
				return errors.New("Zsync: Get interrupt signal")
			case <-running:
				return nil
			}
		}, func(error) {
			close(running)
		})
	logger.Info("Zsync started")
	logger.Info(p.group.Run())
	close(p.quit)
}
