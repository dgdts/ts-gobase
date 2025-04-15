package cron

import (
	"sync"

	cron "github.com/robfig/cron/v3"
)

var c *cron.Cron
var ids map[string]cron.EntryID
var mutex sync.Mutex

func init() {
	c = cron.New()
	c.Start()
	ids = make(map[string]cron.EntryID)
}

type CronJob interface {
	cron.Job
	Symbol() string
}

func RemoveFunc(symbol string) {
	mutex.Lock()
	defer mutex.Unlock()
	id, ok := ids[symbol]
	if !ok {
		return
	}
	c.Remove(id)
}

func Raw() *cron.Cron {
	return c
}
