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
	jour       time.Time `col:"jour" size:"32" dateFormat:"yyyy-mm-dd"`
	username   string    `col:"utilisateur" size:"50"`
	actions    int       `col:"actions" size:"8"`
	recherches int       `col:"recherches" size:"8"`
	fiches     int       `col:"fiches" size:"8"`
	segment    string    `col:"segment" size:"16"`
}

type activiteParJourSelector struct {
	from, to time.Time
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
		item:      activiteParJour{},
		sheetName: "activités par jour",
		asRow:     activiteParJoursToRow,
	}
}

func activiteParJoursToRow(ligne activiteParJour) []any {
	return []any{
		ligne.jour,
		ligne.username,
		ligne.actions,
		ligne.recherches,
		ligne.fiches,
		ligne.segment,
	}
}
