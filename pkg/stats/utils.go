package stats

import (
	"bytes"
	"encoding/csv"

	"github.com/pkg/errors"

	"datapi/pkg/utils"
)

func transformLogsToData(logs []line) ([]byte, error) {
	var data bytes.Buffer
	writer := csv.NewWriter(&data)
	err := utils.Apply(logs, func(l line) error {
		wErr := writer.Write(l.getFieldsAsStringArray())
		if wErr != nil {
			err := errors.Wrap(wErr, "erreur lors de la transformation de la ligne de log "+l.String())
			return err
		}
		return nil
	})
	writer.Flush()
	return data.Bytes(), err
}
