package stats

import (
	"context"
	_ "embed"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
	err      error
}

func selectActiviteParUtilisateur(
	ctx context.Context,
	dbPool *pgxpool.Pool,
	since time.Time,
	to time.Time,
	r chan activiteParUtilisateur,
) {
	defer close(r)
	rows, err := dbPool.Query(ctx, selectActiviteParUtilisateurSQL, since.Truncate(day), to.Truncate(day))
	if err != nil {
		r <- activiteParUtilisateur{err: errors.Wrap(err, "erreur pendant la requête de sélection des activite")}
	}
	for rows.Next() {
		var activite activiteParUtilisateur
		err := rows.Scan(&activite.username, &activite.visites, &activite.actions, &activite.segment)
		if err != nil {
			r <- activiteParUtilisateur{err: errors.Wrap(err, "erreur pendant la récupération des résultats")}
		}
		r <- activite
	}
	if err := rows.Err(); err != nil {
		r <- activiteParUtilisateur{err: errors.Wrap(err, "erreur après la récupération des résultats")}
	}
}

func writeActivitesUtilisateurToExcel(xls *excelize.File, pageIndex int, activites chan activiteParUtilisateur) error {
	var sheetName, err = createSheet(xls, "Activité par utilisateur", pageIndex)
	if err != nil {
		return err
	}
	var row = 1
	if activites != nil {
		for ligne := range activites {
			if ligne.err != nil {
				return ligne.err
			}
			err := writeActiviteUtilisateurToExcel(xls, sheetName, ligne, row)
			if err != nil {
				return err
			}
			row++
		}
	}
	return nil
}

func writeActiviteUtilisateurToExcel(f *excelize.File, sheetName string, ligne activiteParUtilisateur, row int) error {
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
