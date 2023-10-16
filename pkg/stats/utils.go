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

func createExcel(archive *zip.Writer, filename string, activitesParUtilisateur chan row[activiteParUtilisateur], activitesParJour chan activiteParJour) error {
	slog.Debug("crée l'entrée excel", slog.String("filename", filename))
	zipEntry, err := archive.CreateHeader(&zip.FileHeader{
		Name:     filename + ".xlsx",
		Comment:  "fourni par Datapi avec dégoût",
		NonUTF8:  false,
		Modified: time.Now(),
	})
	if err != nil {
		slog.Error("erreur pendant la création du xls dans le zip", slog.Any("error", err))
		return err
	}
	excelFile := newExcel()
	defer func() {
		if err := excelFile.Close(); err != nil {
			slog.Error("erreur à la fermeture du fichier", slog.Any("error", err), slog.String("filename", excelFile.Path))
		}
	}()
	err = writeActivitesUtilisateurToExcel(excelFile, 0, activitesParUtilisateur)
	if err != nil {
		slog.Error("erreur pendant l'écriture des activités utilisateurs dans le xls", slog.Any("error", err))
		return err
	}
	err = writeActivitesJoursToExcel(excelFile, 1, activitesParJour)
	if err != nil {
		slog.Error("erreur pendant l'écriture des activités jour dans le xls", slog.Any("error", err))
		return err
	}
	return exportTo(zipEntry, excelFile)
}
