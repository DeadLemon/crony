package cron

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/DeadLemon/crony/cron/cmd"
	"github.com/DeadLemon/crony/crontab"
)

func StartJob(ctx context.Context, tab *crontab.Context, job *crontab.Job, log *zap.Logger) {
	var wg sync.WaitGroup
	defer wg.Wait()

	nextRun := time.Now()

	for {
		nextRun = job.Expression.Next(nextRun)
		delay := nextRun.Sub(time.Now())
		if delay < 0 {
			nextRun = time.Now()
			continue
		}

		log.Info("scheduled", zap.Time("next", nextRun))

		select {
		case <-ctx.Done():
			log.Info("shutting down")
			return
		case <-time.After(delay):
			log.Info("running")
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd.New(tab, job).Run()
		}()
	}
}
