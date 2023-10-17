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

//go:embed resources/sql/select_activite_par_jour.sql
var selectActiviteParJourSQL string

type activiteParJour struct {
	jour       time.Time
	username   string
	actions    int
	recherches int
	fiches     int
	segment    string
	err        error
}

func selectActiviteParJour(ctx context.Context, dbPool *pgxpool.Pool, since time.Time, to time.Time, r chan row[activiteParJour]) {
	defer close(r)
	rows, err := dbPool.Query(ctx, selectActiviteParJourSQL, since.Truncate(day), to.Truncate(day))
	if err != nil {
		r <- rowWithError(activiteParJour{}, errors.Wrap(err, "erreur pendant la requête de sélection des activite/jour"))
	}
	for rows.Next() {
		var activite activiteParJour
		err := rows.Scan(
			&activite.jour,
			&activite.username,
			&activite.actions,
			&activite.recherches,
			&activite.fiches,
			&activite.segment,
		)
		if err != nil {
			r <- rowWithError(activiteParJour{}, errors.Wrap(err, "erreur pendant la récupération des résultats/jour"))
		}
		r <- newRow(activite)
	}
	if err := rows.Err(); err != nil {
		r <- rowWithError(activiteParJour{}, errors.Wrap(err, "erreur après la récupération des résultats/jour"))
	}
}

func writeOneActiviteJourToExcel(f *excelize.File, sheetName string, ligne activiteParJour, row int) error {
	var i = 1
	err := writeString(f, sheetName, ligne.jour.Format(time.DateOnly), i, row)
	i++
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du jour")
	}
	err = writeString(f, sheetName, ligne.username, i, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du nom de l'utilisateur")
	}
	i++
	err = writeInt(f, sheetName, ligne.actions, i, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du nombre d'actions")
	}
	i++
	err = writeInt(f, sheetName, ligne.recherches, i, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du nombre de recherches")
	}
	i++
	err = writeInt(f, sheetName, ligne.fiches, i, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du nombre de fiches")
	}
	i++
	err = writeString(f, sheetName, ligne.segment, i, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du nombre de segments")
	}
	return nil
}

func writeActivitesJoursToExcel(xls *excelize.File, pageIndex int, activites chan row[activiteParJour]) error {
	err := writeOneSheetToExcel(xls, "Activité par jour", pageIndex, activites, writeOneActiviteJourToExcel)
	if err != nil {
		return fmt.Errorf("erreur lors de l'écriture d'une ligne d'activités par utilisateurs : %w", err)
	}
	return nil
}
