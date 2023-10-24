package stats

import (
	"sort"
	"strings"
)

const accessLogDateLayout = "20060102150405"

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

type row[A any] struct {
	value A
	err   error
}

func rowWithError[A any](_ A, err error) row[A] {
	return row[A]{err: err}
}

func newRow[A any](value A) row[A] {
	return row[A]{value: value}
}
