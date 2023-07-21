package stats

import (
	"strings"
	"time"
)

const AccessLogDateLayout = "20060102150405"

type accessLog struct {
	date     time.Time
	path     string
	method   string
	username string
	roles    []string
	err      error
}

func (l accessLog) String() string {
	return strings.Join(l.toStringArray(), ";")
}

func (l accessLog) toStringArray() []string {
	return []string{
		l.date.Format(AccessLogDateLayout),
		l.path,
		l.method,
		l.username,
		strings.Join(l.roles, "-"),
	}
}
