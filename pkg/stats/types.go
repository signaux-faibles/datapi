package stats

import (
	"sort"
	"strings"
	"time"
)

const accessLogDateLayout = "20060102150405"

type accessLog struct {
	date     time.Time
	path     string
	method   string
	username string
	segment  string
	roles    []string
	err      error
}

func (l accessLog) String() string {
	return strings.Join(l.toStringArray(), ";")
}

func (l accessLog) toStringArray() []string {
	sort.Strings(l.roles)
	return []string{
		l.date.Format(accessLogDateLayout),
		l.path,
		l.method,
		l.username,
		l.segment,
		strings.Join(l.roles, ","),
	}
}