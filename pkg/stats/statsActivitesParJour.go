package stats

import (
	_ "embed"
	"time"

	"github.com/jackc/pgx/v5"
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

func writeActiviteJourHeaders(f *excelize.File, sheetName string) {
	options := excelize.HeaderFooterOptions{
		AlignWithMargins: true,
		DifferentFirst:   false,
		DifferentOddEven: true,
		ScaleWithDoc:     false,
		OddHeader:        "OH",
		OddFooter:        "OF",
		EvenHeader:       "EH",
		EvenFooter:       "EF",
		FirstHeader:      "FH",
		FirstFooter:      "FF",
	}
	f.SetHeaderFooter(sheetName, &options)
}
