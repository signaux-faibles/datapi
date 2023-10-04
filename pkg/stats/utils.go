package stats

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
)

const accessLogDateExcelLayout = "2006/01/02 15:04:05"

var csvHeaders = []string{"date", "chemin", "methode", "utilisateur", "segment", "roles"}

func writeLinesToCSV(logs chan accessLog, w io.Writer) error {
	csvWriter := csv.NewWriter(w)
	counter := 0
	wErr := csvWriter.Write(csvHeaders)
	if wErr != nil {
		err := errors.Wrap(wErr, "erreur lors de l'écriture des headers")
		return err
	}
	for l := range logs {
		if l.err != nil {
			return l.err
		}
		wErr := csvWriter.Write(l.toCSV())
		if wErr != nil {
			err := errors.Wrap(wErr, "erreur lors de la transformation de la ligne de log")
			return err
		}
		counter++
	}
	csvWriter.Flush()
	return nil
}

func writeToExcel(file io.Writer) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	// Create a new sheet.
	index, err := f.NewSheet("Sheet2")
	if err != nil {
		fmt.Println(err)
		return err
	}
	// Set value of a cell.
	err = f.SetCellValue("Sheet2", "A2", "Hello world.")
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = f.SetCellValue("Sheet1", "B2", 100)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// Set active sheet of the workbook.
	f.SetActiveSheet(index)
	_, err = f.WriteTo(file, excelize.Options{RawCellValue: true})
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (l accessLog) toCSV() []string {
	sort.Strings(l.roles)
	return []string{
		l.date.Format(accessLogDateExcelLayout),
		l.path,
		l.method,
		l.username,
		l.segment,
		strings.Join(l.roles, ","),
	}
}
