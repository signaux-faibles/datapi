package stats

import (
	"strings"
	"time"
)

type line struct {
	date     time.Time
	path     string
	method   string
	username string
	roles    []string
}

func (l line) String() string {
	return strings.Join(l.getFieldsAsStringArray(), ";")
}
