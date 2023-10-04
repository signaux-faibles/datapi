package stats

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
)

const accessLogDateExcelLayout = "2006/01/02 15:04:05"

var csvHeaders = []string{"date", "chemin", "methode", "utilisateur", "segment", "roles"}

func writeLinesToCSV(logs chan accessLog, w io.Writer) error {
	csvWriter := csv.NewWriter(w)
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

func createCSV(archive *zip.Writer, filename string, results chan accessLog) error {
	csvFile, err := archive.CreateHeader(&zip.FileHeader{
		Name:     filename + ".csv",
		Comment:  "fourni par Datapi avec amour",
		NonUTF8:  false,
		Modified: time.Now(),
	})
	if err != nil {
		slog.Error("erreur pendant la création du csv dans le zip", slog.Any("error", err))
		return err
	}
	err = writeLinesToCSV(results, csvFile)
	if err != nil {
		slog.Error("erreur pendant l'écriture du csv", slog.Any("error", err))
		return err
	}
	return nil
}

func createExcel(archive *zip.Writer, filename string) error {
	excelFile, err := archive.CreateHeader(&zip.FileHeader{
		Name:     filename + ".xslx",
		Comment:  "fourni par Datapi avec dégoût",
		NonUTF8:  false,
		Modified: time.Now(),
	})
	if err != nil {
		slog.Error("erreur pendant la création du xls dans le zip", slog.Any("error", err))
		return err
	}
	err = writeToExcel(excelFile)
	if err != nil {
		slog.Error("erreur pendant l'écriture du xls", slog.Any("error", err))
		return err
	}
	return nil
}
