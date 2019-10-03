package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/DeadLemon/crony/cron"
	"github.com/DeadLemon/crony/crontab"
)

func New() *cobra.Command {
	return &cobra.Command{
		Use: "cron",
		Run: func(cmd *cobra.Command, args []string) {
			log, err := zap.NewDevelopment()
			if err != nil {
				return
			}

			tab, err := crontab.New(args[0])
			if err != nil {
				log.Error("failed to read crontab", zap.Error(err))
				return
			}

			var wg sync.WaitGroup

			ctx, cancel := context.WithCancel(context.Background())

			for _, job := range tab.Jobs {
				wg.Add(1)
				go func(job *crontab.Job) {
					defer wg.Done()
					cron.StartJob(
						ctx, tab.Context, job,
						log.Named("job").With(
							zap.String("schedule", job.Schedule),
							zap.String("command", job.Command),
						),
					)
				}(job)
			}

			termChan := make(chan os.Signal, 1)
			signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

			termSig := <-termChan

			log.Info(fmt.Sprintf("received %s, shutting down", termSig))
			cancel()

			log.Info("waiting for jobs to finish")
			wg.Wait()

			log.Info("exiting")
		},
	}
}
