package stats

import (
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
)

func newExcel() *excelize.File {
	return excelize.NewFile()
}

//func createSheet(f *excelize.File, sheetName string, index int) (string, error) {
//	// Create a new sheet.
//	originalName := f.GetSheetName(index)
//	if originalName == "" {
//		_, err := f.NewSheet(sheetName)
//		if err != nil {
//			return "", err
//		}
//		return sheetName, nil
//	}
//	err := f.SetSheetName(originalName, sheetName)
//	if err != nil {
//		return "", err
//	}
//	// f.SetActiveSheet(index)
//	return sheetName, err
//}

func addSheet(f *excelize.File, sheetName string) (int, error) {
	defaultSheetName := "Sheet1"
	if len(f.GetSheetList()) == 1 && f.GetSheetName(0) == defaultSheetName {
		err := f.SetSheetName(defaultSheetName, sheetName)
		if err != nil {
			return -1, errors.Wrap(err, "erreur lors du changement de nom de la première feuille")
		}
		return f.GetSheetIndex(sheetName)
	}
	return f.NewSheet(sheetName)
}

type ExcelRowWriter[Row any] func(f *excelize.File, sheetName string, ligne Row, row int) error

func writeOneSheetToExcel[A any](
	xls *excelize.File,
	sheetLabel string,
	itemsToWrite chan row[A],
	writer ExcelRowWriter[A],
) error {
	_, err := addSheet(xls, sheetLabel)
	if err != nil {
		return err
	}
	var row = 1
	if itemsToWrite != nil {
		for ligne := range itemsToWrite {
			if ligne.err != nil {
				return ligne.err
			}
			err := writer(xls, sheetLabel, ligne.value, row)
			if err != nil {
				return err
			}
			row++
		}
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

func writeStrings(f *excelize.File, sheetName string, value []string, col, row int) error {
	return writeString(f, sheetName, strings.Join(value, "; "), col, row)
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
