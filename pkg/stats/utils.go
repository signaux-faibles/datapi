package stats

import (
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"datapi/pkg/utils"
)

func writeLinesToCompressedCSV(logs chan accessLog, maxResults int) ([]byte, error) {
	var data bytes.Buffer
	var err error
	zipWriter := gzip.NewWriter(&data)
	csvWriter := csv.NewWriter(zipWriter)
	counter := 0
	for l := range logs {
		if l.err != nil {
			return nil, l.err
		}
		wErr := csvWriter.Write(l.getFieldsAsStringArray())
		if wErr != nil {
			err := errors.Wrap(wErr, "erreur lors de la transformation de la ligne de log "+l.String())
			return nil, err
		}
		counter++
	}
	if counter == 0 {
		return nil, utils.NewJSONerror(http.StatusBadRequest, "aucune logs trouvée, les critères sont trop restrictifs")
	}
	if counter > maxResults {
		return nil, utils.NewJSONerror(http.StatusBadRequest, "trop de logs trouvées, les critères ne sont pas assez restrictifs")
	}
	csvWriter.Flush()
	err = zipWriter.Flush()
	if err != nil {
		return nil, errors.Wrap(err, "erreur lors de l'écriture du zip")
	}
	if closeErr := zipWriter.Close(); closeErr != nil {
		err = errors.Wrap(closeErr, "erreur lors de la fermeture du zip")
	}
	return data.Bytes(), err
}

func writeLinesToCSV(logs chan accessLog, maxResults int, w io.Writer) error {
	var err error
	csvWriter := csv.NewWriter(w)
	counter := 0
	for l := range logs {
		if l.err != nil {
			return l.err
		}
		wErr := csvWriter.Write(l.getFieldsAsStringArray())
		if wErr != nil {
			err := errors.Wrap(wErr, "erreur lors de la transformation de la ligne de log "+l.String())
			return err
		}
		counter++
	}
	if counter == 0 {
		return utils.NewJSONerror(http.StatusBadRequest, "aucune logs trouvée, les critères sont trop restrictifs")
	}
	if counter > maxResults {
		return utils.NewJSONerror(http.StatusBadRequest, "trop de logs trouvées, les critères ne sont pas assez restrictifs")
	}
	csvWriter.Flush()
	return err
}

//func closeZip(archive *zip.Writer, err error) {
//  if archive != nil {
//    defer func() {
//      if err := archive.Close(); err != nil {
//        utils.AbortWithError(c, err)
//      }
//    }()
//  }
//}
