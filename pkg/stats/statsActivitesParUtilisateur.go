package stats

import (
	"context"
	_ "embed"
	"fmt"
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
}

func (a activiteParUtilisateur) row() row[activiteParUtilisateur] {
	return row[activiteParUtilisateur]{value: a}
}

func selectActiviteParUtilisateur(
	ctx context.Context,
	dbPool *pgxpool.Pool,
	since time.Time,
	to time.Time,
	r chan row[activiteParUtilisateur],
) {
	defer close(r)
	rows, err := dbPool.Query(ctx, selectActiviteParUtilisateurSQL, since.Truncate(day), to.Truncate(day))
	if err != nil {
		r <- rowWithError(activiteParUtilisateur{}, errors.Wrap(err, "erreur pendant la requête de sélection des activite"))
	}
	for rows.Next() {
		var activite activiteParUtilisateur
		err := rows.Scan(&activite.username, &activite.visites, &activite.actions, &activite.segment)
		if err != nil {
			r <- rowWithError(activiteParUtilisateur{}, errors.Wrap(err, "erreur pendant la récupération des résultats"))
		}
		r <- activite.row()
	}
	if err := rows.Err(); err != nil {
		r <- rowWithError(activiteParUtilisateur{}, errors.Wrap(err, "erreur après la récupération des résultats"))
	}
}

func writeActivitesUtilisateurToExcel(xls *excelize.File, pageIndex int, activites chan row[activiteParUtilisateur]) error {
	err := writeOneSheetToExcel(xls, "Activité par utilisateur", pageIndex, activites, writeOneActiviteUtilisateurToExcel)
	if err != nil {
		return fmt.Errorf("erreur lors de l'écriture d'une ligne d'activités par utilisateurs : %w", err)
	}
	return nil
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
