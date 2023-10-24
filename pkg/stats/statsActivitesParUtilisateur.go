package stats

import (
	_ "embed"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

//go:embed resources/sql/select_activite_par_utilisateur.sql
var selectActiviteParUtilisateurSQL string

type activiteParUtilisateur struct {
	username string
	actions  string
	visites  string
	segment  string
}

type activiteParUtilisateurSelector struct {
	from, to time.Time
}

var activitesParUtilisateurHeaders = map[any]float64{
	"utilisateur": float64(-1),
	"actions":     float64(-1),
	"visites":     float64(-1),
	"segment":     float64(-1),
}

func newActiviteParUtilisateurSelector(from time.Time, to time.Time) activiteParUtilisateurSelector {
	return activiteParUtilisateurSelector{from: from.Truncate(day), to: to.Truncate(day)}
}

func (a activiteParUtilisateurSelector) sql() string {
	return selectActiviteParUtilisateurSQL
}

func (a activiteParUtilisateurSelector) sqlArgs() []any {
	return []any{a.from, a.to}
}

func (a activiteParUtilisateurSelector) toItem(rows pgx.Rows) (activiteParUtilisateur, error) {
	var r activiteParUtilisateur
	err := rows.Scan(&r.username, &r.visites, &r.actions, &r.segment)
	if err != nil {
		return activiteParUtilisateur{}, errors.Wrap(err, "erreur lors la création de l'objet")
	}
	return r, nil
}

func activitesUtilisateurSheetConfig() sheetConfig[activiteParUtilisateur] {
	return anySheetConfig[activiteParUtilisateur]{
		sheetName:      "activités par utilisateur",
		headersAndSize: activitesParUtilisateurHeaders,
		startRow:       3,
		asRow:          activiteParUtilisateurToRow,
	}
}

func activiteParUtilisateurToRow(ligne activiteParUtilisateur) []any {
	return []any{
		ligne.username,
		ligne.actions,
		ligne.visites,
		ligne.segment,
	}
}
