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

type accessLog struct {
	date     time.Time `col:"date" size:"36" dateFormat:"yyyy-mm-dd hh:mm:ss"`
	path     string    `col:"chemin" size:"36"`
	method   string    `col:"méthode" size:"10"`
	username string    `col:"utilisateur" size:"36"`
	segment  string    `col:"segment" size:"24"`
	roles    []string  `col:"rôles" size:"100"`
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
		item:      accessLog{},
		sheetName: "access logs",
		asRow:     toRow,
	}
}

func toRow(ligne accessLog) []any {
	return []any{
		ligne.date,
		ligne.path,
		ligne.method,
		ligne.username,
		ligne.segment,
		strings.Join(ligne.roles, "; "),
	}
}
