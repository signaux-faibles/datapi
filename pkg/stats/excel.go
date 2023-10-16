package stats

import (
	"io"
	"log/slog"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
)

func newExcel() *excelize.File {
	return excelize.NewFile()
}

func createSheet(f *excelize.File, sheetName string, index int) (string, error) {
	// Create a new sheet.
	originalName := f.GetSheetName(index)
	if originalName == "" {
		_, err := f.NewSheet(sheetName)
		if err != nil {
			return "", err
		}
		return sheetName, nil
	}
	err := f.SetSheetName(originalName, sheetName)
	if err != nil {
		return "", err
	}
	// f.SetActiveSheet(index)
	return sheetName, err
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

func writeActiviteJourToExcel(f *excelize.File, sheetName string, ligne activiteParJour, row int) error {
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

func writeString(f *excelize.File, sheetName string, value string, col, row int) error {
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant la récupération des coordonnées")
	}
	err = f.SetCellStr(sheetName, cell, value)
	return err
}

func writeInt(f *excelize.File, sheetName string, value int, col, row int) error {
	return writeString(f, sheetName, strconv.Itoa(value), col, row)
}

func exportTo(output io.Writer, xlsx *excelize.File) error {
	err := xlsx.Write(output, excelize.Options{RawCellValue: true})
	if err != nil {
		slog.Error("erreur pendant l'écriture du fichier xlsx")
		return err
	}
	return nil
}

func writeActivitesJoursToExcel(xls *excelize.File, pageIndex int, activites chan activiteParJour) error {
	var sheetName, err = createSheet(xls, "Activité par jour", pageIndex)
	if err != nil {
		return errors.Wrap(err, "erreur lors de la création de la feuille excel")
	}
	var row = 1
	if activites != nil {
		for ligne := range activites {
			if ligne.err != nil {
				return ligne.err
			}
			err := writeActiviteJourToExcel(xls, sheetName, ligne, row)
			if err != nil {
				return err
			}
			row++
		}
	}
	return nil
}
