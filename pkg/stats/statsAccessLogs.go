package stats

import (
	_ "embed"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
)

const accessLogTitle = "Access logs"

//go:embed resources/sql/select_logs.sql
var selectLogsSQL string

func (ligne accessLog) toRow() *[]any {
	return &[]any{
		ligne.date.Format(time.DateTime),
		ligne.path,
		ligne.method,
		ligne.username,
		ligne.segment,
		strings.Join(ligne.roles, "; "),
	}
}

func writeOneAccessLogToExcel(f *excelize.File, sheetName string, ligne accessLog, row int) error {
	cellName, err := excelize.CoordinatesToCellName(1, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant la récupération du nom de la cellule")
	}
	err = f.SetSheetRow(sheetName, cellName, ligne.toRow())
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture de la ligne'")
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
