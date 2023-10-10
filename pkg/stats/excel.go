package stats

import (
	"io"
	"log/slog"

	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
)

func newExcel() *excelize.File {
	return excelize.NewFile()
}

func createSheet(f *excelize.File, sheetName string, index int) (string, error) {
	// Create a new sheet.
	err := f.SetSheetName(f.GetSheetName(index), sheetName)
	if err != nil {
		return "", err
	}
	f.SetActiveSheet(index)
	return sheetName, err
}

func writeActiviteToExcel(f *excelize.File, sheetName string, ligne activiteParUtilisateur, row int) error {
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

func writeString(f *excelize.File, sheetName string, value string, col, row int) error {
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant la récupération des coordonnées")
	}
	err = f.SetCellStr(sheetName, cell, value)
	return err
}

func exportTo(output io.Writer, xlsx *excelize.File) error {
	err := xlsx.Write(output, excelize.Options{RawCellValue: true})
	if err != nil {
		slog.Error("erreur pendant l'écriture du fichier xlsx")
		return err
	}
	return nil
}
