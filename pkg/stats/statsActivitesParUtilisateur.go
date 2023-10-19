package stats

import (
	_ "embed"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
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

func writeOneActiviteUtilisateurToExcel(f *excelize.File, sheetName string, ligne activiteParUtilisateur, row int) error {
	err := writeString(f, sheetName, ligne.username, 1, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du username")
	}
	err = writeString(f, sheetName, ligne.visites, 2, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du nombre de visites")
	}
	err = writeString(f, sheetName, ligne.actions, 3, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du nombre d'actions")
	}
	err = writeString(f, sheetName, ligne.segment, 4, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du nombre de segments")
	}
	return nil
}
