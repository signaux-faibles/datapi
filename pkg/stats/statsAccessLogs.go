package stats

import (
	_ "embed"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

//go:embed resources/sql/select_logs.sql
var selectLogsSQL string

var accessLogHeaders = map[any]float64{
	"date":     float64(-1),
	"chemin":   float64(-1),
	"méthode":  float64(-1),
	"username": float64(-1),
	"segment":  float64(-1),
	"rôle":     float64(-1),
}

type accessLogsSelector struct {
	from time.Time
	to   time.Time
}

func newAccessLogsSelector(from time.Time, to time.Time) accessLogsSelector {
	return accessLogsSelector{from: from.Truncate(day), to: to.Truncate(day)}
}

func (a accessLogsSelector) sql() string {
	return selectLogsSQL
}

func (a accessLogsSelector) sqlArgs() []any {
	return []any{a.from, a.to}
}

func (a accessLogsSelector) toItem(rows pgx.Rows) (accessLog, error) {
	var r accessLog
	err := rows.Scan(&r.date, &r.path, &r.method, &r.username, &r.segment, &r.roles)
	if err != nil {
		return accessLog{}, errors.Wrap(err, "erreur lors la création de l'objet")
	}
	return r, nil
}

func accessLogSheetConfig() sheetConfig[accessLog] {
	return anySheetConfig[accessLog]{
		sheetName:      "access logs",
		headersAndSize: accessLogHeaders,
		startRow:       3,
		asRow:          toRow,
	}
}

func toRow(ligne accessLog) []any {
	return []any{
		ligne.date.Format(time.DateTime),
		ligne.path,
		ligne.method,
		ligne.username,
		ligne.segment,
		strings.Join(ligne.roles, "; "),
	}
}
