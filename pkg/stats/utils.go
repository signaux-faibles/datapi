package stats

import (
	"archive/zip"
	"encoding/csv"
	"io"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
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

func writeToExcel(output io.Writer, activites chan activiteParUtilisateur) error {
	f := newExcel()
	defer func() {
		if err := f.Close(); err != nil {
			slog.Error("erreur à la fermeture du fichier", slog.Any("error", err), slog.String("filename", f.Path))
		}
	}()
	var sheetName, err = createSheet(f, "Activité par utilisateur", 0)
	if err != nil {
		return err
	}
	var row = 1
	if activites != nil {
		for ligne := range activites {
			if ligne.err != nil {
				return ligne.err
			}
			err := writeActiviteToExcel(f, sheetName, ligne, row)
			if err != nil {
				return err
			}
			row++
		}
	}
	return exportTo(output, f)
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

func createExcel(archive *zip.Writer, filename string, activites chan activiteParUtilisateur) error {
	slog.Info("crée l'entrée excel", slog.String("filename", filename))
	excelFile, err := archive.CreateRaw(&zip.FileHeader{
		Name:     filename + ".xlsx",
		Comment:  "fourni par Datapi avec dégoût",
		NonUTF8:  false,
		Modified: time.Now(),
	})
	if err != nil {
		slog.Error("erreur pendant la création du xls dans le zip", slog.Any("error", err))
		return err
	}
	err = writeToExcel(excelFile, activites)
	if err != nil {
		slog.Error("erreur pendant l'écriture du xls", slog.Any("error", err))
		return err
	}
	return nil
}
