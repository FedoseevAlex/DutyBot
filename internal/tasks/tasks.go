package tasks

import (
	"time"

	"github.com/robfig/cron/v3"
)

var scheduler *cron.Cron

func InitScheduler() {
	scheduler = cron.New(cron.WithLocation(time.UTC))
}

func Start() {
	scheduler.Start()
}

func Stop() {
	scheduler.Stop()
}

func AddTask(period string, job func()) (cron.EntryID, error) {
	entryID, err := scheduler.AddFunc(period, job)
	if err != nil {
		return -1, err
	}
	return entryID, nil
}
