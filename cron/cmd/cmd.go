package cmd

import (
	"fmt"
	"os"
	"time"

	exec "github.com/go-cmd/cmd"

	"github.com/DeadLemon/crony/crontab"
)

type Cmd struct {
	*exec.Cmd
}

func New(tab *crontab.Context, job *crontab.Job) *Cmd {
	options := exec.Options{Buffered: false, Streaming: true}
	r := exec.NewCmdOptions(options, tab.Shell, "-c", job.Command)

	env := os.Environ()
	for k, v := range tab.Environ {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	r.Env = env

	return &Cmd{r}
}

func (r *Cmd) Run() {
	go func() {
		for {
			select {
			case line := <-r.Stdout:
				_, _ = fmt.Fprintln(os.Stdout, line)
			case line := <-r.Stderr:
				_, _ = fmt.Fprintln(os.Stderr, line)
			}
		}
	}()

	<-r.Start()

	for len(r.Stdout) > 0 || len(r.Stderr) > 0 {
		time.Sleep(10 * time.Millisecond)
	}
}
