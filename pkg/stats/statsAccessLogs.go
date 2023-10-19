package stats

import (
	_ "embed"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
)

const accessLogTitle = "Access logs"

//go:embed resources/sql/select_logs.sql
var selectLogsSQL string

func writeOneAccessLogToExcel(f *excelize.File, sheetName string, ligne accessLog, row int) error {
	var i = 0
	i++
	err := writeString(f, sheetName, ligne.date.Format(time.DateTime), i, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture de la date")
	}
	i++
	err = writeString(f, sheetName, ligne.path, i, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du chemin")
	}
	i++
	err = writeString(f, sheetName, ligne.method, i, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du verbe HTTP")
	}
	i++
	err = writeString(f, sheetName, ligne.username, i, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du nom de l'utilisateur")
	}
	i++
	err = writeString(f, sheetName, ligne.segment, i, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du segment")
	}
	i++
	err = writeStrings(f, sheetName, ligne.roles, i, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du nombre de segments")
	}
	return nil
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
