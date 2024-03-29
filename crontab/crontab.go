package crontab

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/gorhill/cronexpr"
)

var (
	jobLineSeparator = regexp.MustCompile(`\S+`)
	envLineMatcher   = regexp.MustCompile(`^([^\s=]+)\s*=\s*(.*)$`)

	parameterCounts = []int{
		7, // POSIX + seconds + years
		6, // POSIX + years
		5, // POSIX
		1, // shorthand (e.g. @hourly)
	}
)

func parseJobLine(line string) (*Line, error) {
	indices := jobLineSeparator.FindAllStringIndex(line, -1)

	for _, count := range parameterCounts {
		if len(indices) <= count {
			continue
		}

		scheduleEnds := indices[count-1][1]
		commandStarts := indices[count][0]

		// TODO: Should receive a logger?
		log.Println("try parse(%d): %s[0:%d] = %s", count, line, scheduleEnds, line[0:scheduleEnds])

		expr, err := cronexpr.Parse(line[:scheduleEnds])

		if err != nil {
			continue
		}

		return &Line{
			Expression: expr,
			Schedule:   line[:scheduleEnds],
			Command:    line[commandStarts:],
		}, nil
	}
	return nil, fmt.Errorf("bad crontab line: %s", line)
}

func parseCrontab(reader io.Reader) (*Crontab, error) {
	scanner := bufio.NewScanner(reader)

	position := 0

	jobs := make([]*Job, 0)

	// TODO: CRON_TZ?
	environ := make(map[string]string)
	shell := "/bin/sh"

	for scanner.Scan() {
		line := strings.TrimLeft(scanner.Text(), " \t")

		if line == "" {
			continue
		}

		if line[0] == '#' {
			continue
		}

		r := envLineMatcher.FindAllStringSubmatch(line, -1)
		if len(r) == 1 && len(r[0]) == 3 {
			envKey := r[0][1]
			envVal := r[0][2]

			// Remove quotes (this emulates what Vixie cron does)
			if envVal[0] == '"' || envVal[0] == '\'' {
				if len(envVal) > 1 && envVal[0] == envVal[len(envVal)-1] {
					envVal = envVal[1 : len(envVal)-1]
				}
			}

			if envKey == "SHELL" {
				log.Println("processes will be spawned using shell: %s", envVal)
				shell = envVal
			}

			if envKey == "USER" {
				log.Println("processes will NOT be spawned as USER=%s", envVal)
			}

			environ[envKey] = envVal

			continue
		}

		jobLine, err := parseJobLine(line)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, &Job{Line: *jobLine, Position: position})
		position++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &Crontab{
		Jobs: jobs,
		Context: &Context{
			Shell:   shell,
			Environ: environ,
		},
	}, nil
}

func New(path string) (*Crontab, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	return parseCrontab(file)
}
