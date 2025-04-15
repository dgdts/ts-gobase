package cron

import (
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/robfig/cron/v3"
)

// tip: if last job was not finished, it will be delayed
func AddCronFunc(symbol string, spec string, f func()) error {
	mutex.Lock()
	defer mutex.Unlock()
	_, ok := ids[symbol]
	if ok {
		return fmt.Errorf("%s cron job already exists", symbol)
	}
	var logJob = cron.FuncJob(func() {
		start := time.Now()
		f()
		cost := time.Since(start).Seconds()
		if cost > 10 {
			hlog.Warnf("cron job %s cost:%.2fs", symbol, cost)
		}
	})
	id, err := c.AddJob(spec, cron.NewChain(cron.DelayIfStillRunning(cron.DefaultLogger)).Then(logJob))
	if err != nil {
		return err
	}
	ids[symbol] = id
	return nil
}

func AddFunc(symbol string, every time.Duration, f func()) error {
	s := every.Seconds()
	if s < 1 {
		s = 1
	}
	spec := fmt.Sprintf("@every %.0fs", s)
	return AddCronFunc(symbol, spec, f)
}

func AddJob(every time.Duration, job CronJob) error {
	hlog.Infof("cron job:%s, every:%v second", job.Symbol(), every.Seconds())
	return AddFunc(job.Symbol(), every, job.Run)
}

func MustAddJob(every time.Duration, job CronJob) {
	err := AddJob(every, job)
	if err != nil {
		panic(fmt.Sprintf("add cron job err:%v", err))
	}
}

func MustAddCronFunc(spec string, job CronJob) {
	err := AddCronFunc(job.Symbol(), spec, job.Run)
	if err != nil {
		panic(fmt.Sprintf("add cron job err:%v", err))
	}
}
