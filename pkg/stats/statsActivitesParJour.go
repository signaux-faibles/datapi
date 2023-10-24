package stats

import (
	_ "embed"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

//go:embed resources/sql/select_activite_par_jour.sql
var selectActiviteParJourSQL string

type activiteParJour struct {
	jour       time.Time
	username   string
	actions    int
	recherches int
	fiches     int
	segment    string
}

type activiteParJourSelector struct {
	from, to time.Time
}

var activitesParJourHeaders = map[any]float64{
	"jour":        float64(-1),
	"utilisateur": float64(-1),
	"actions":     float64(-1),
	"recherches":  float64(-1),
	"fiches":      float64(-1),
	"segment":     float64(-1),
}

func (a activiteParJourSelector) sql() string {
	return selectActiviteParJourSQL
}

func (a activiteParJourSelector) sqlArgs() []any {
	return []any{a.from, a.to}
}

func (a activiteParJourSelector) toItem(rows pgx.Rows) (activiteParJour, error) {
	var r activiteParJour
	err := rows.Scan(&r.jour, &r.username, &r.actions, &r.recherches, &r.fiches, &r.segment)
	if err != nil {
		return activiteParJour{}, errors.Wrap(err, "erreur lors la création de l'objet activiteParJour")
	}
	return r, nil
}

func newActiviteParJourSelector(from time.Time, to time.Time) activiteParJourSelector {
	return activiteParJourSelector{from: from.Truncate(day), to: to.Truncate(day)}
}

func activitesJourSheetConfig() sheetConfig[activiteParJour] {
	return anySheetConfig[activiteParJour]{
		sheetName:      "activités par jour",
		headersAndSize: activitesParJourHeaders,
		startRow:       3,
		asRow:          activiteParJoursToRow,
	}
}

func activiteParJoursToRow(ligne activiteParJour) []any {
	return []any{
		ligne.jour.Format(time.DateOnly),
		ligne.username,
		ligne.actions,
		ligne.recherches,
		ligne.fiches,
		ligne.segment,
	}
}
