package stats

import (
	"bytes"
	"compress/gzip"
	"encoding/csv"

	"github.com/pkg/errors"

	"datapi/pkg/utils"
)

func transformLogsToCompressedData(logs []line) ([]byte, error) {
	var data bytes.Buffer
	zipWriter := gzip.NewWriter(&data)
	csvWriter := csv.NewWriter(zipWriter)
	var err = utils.Apply(logs, func(l line) error {
		wErr := csvWriter.Write(l.getFieldsAsStringArray())
		if wErr != nil {
			err := errors.Wrap(wErr, "erreur lors de la transformation de la ligne de log "+l.String())
			return err
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "erreur lors des lignes de stats dans le zip")
	}
	csvWriter.Flush()
	err = zipWriter.Flush()
	if err != nil {
		return nil, errors.Wrap(err, "erreur lors de l'écriture du zip")
	}
	err = zipWriter.Close()
	if err != nil {
		return nil, errors.Wrap(err, "erreur lors de la fermeture du zip")
	}
	return data.Bytes(), err
}
