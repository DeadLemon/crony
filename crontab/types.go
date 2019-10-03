package crontab

import (
	"time"
)

type Expression interface {
	Next(fromTime time.Time) time.Time
}

type Line struct {
	Expression Expression
	Schedule   string
	Command    string
}

type Job struct {
	Line
	Position int
}

type Context struct {
	Shell   string
	Environ map[string]string
}

type Crontab struct {
	Jobs    []*Job
	Context *Context
}
