package tasks

import (
	"dutybot/internal/config"
	"fmt"
	"log"
	"time"
)

type Task struct {
	StartTime time.Time
	doneChan  chan bool
	Period    time.Duration
	Ticker    *time.Ticker
	Job       func() error
}

func NewTask(job func() error, period time.Duration, startAt time.Time) *Task {
	t := &Task{
		StartTime: startAt,
		Period:    period,
		Job:       job,
		doneChan:  make(chan bool),
	}
	return t
}

func getDurationUntil(point time.Time) time.Duration {
	now := time.Now()
	target := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		point.Hour(),
		point.Minute(),
		point.Second(),
		0,
		now.Location())

	if target.Before(now) {
		target = target.Add(config.Cfg.DutyCycle)
	}
	fmt.Println("Will start on", target)
	return time.Until(target)
}

func (t *Task) Start() {
	f := func() {
		dur := getDurationUntil(t.StartTime)
		<-time.After(dur)
		t.Ticker = time.NewTicker(t.Period)
		for {
			err := t.Job()
			if err != nil {
				log.Fatal("Goroutine failed with:", err)
			}

			select {
			case stop := <-t.doneChan:
				if stop {
					fmt.Println("Task was stopped")
					return
				}
			case <-t.Ticker.C:
				continue
			}
		}
	}
	go f()
}

func (t *Task) Stop() {
	if t.Ticker != nil {
		t.Ticker.Stop()
	}
	t.doneChan <- true
}
