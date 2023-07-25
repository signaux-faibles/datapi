package stats

import (
	"encoding/csv"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"datapi/pkg/utils"
)

const accessLogDateExcelLayout = "2006/01/02 15:04:05"

var csvHeaders = []string{"date", "chemin", "methode", "utilisateur", "roles"}

func writeLinesToCSV(logs chan accessLog, maxResults int, w io.Writer) error {
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
			err := errors.Wrap(wErr, "erreur lors de la transformation de la ligne de log "+l.String())
			return err
		}
		counter++
	}
	csvWriter.Flush()
	if counter == 0 {
		return utils.NewJSONerror(http.StatusBadRequest, "aucune logs trouvée, les critères sont trop restrictifs")
	}
	if counter > maxResults {
		return utils.NewJSONerror(http.StatusBadRequest, "trop de logs trouvées, les critères ne sont pas assez restrictifs")
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
		strings.Join(l.roles, ","),
	}
}
