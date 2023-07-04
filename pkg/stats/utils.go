package stats

import (
	"bytes"
	"encoding/csv"

	"github.com/pkg/errors"

	"datapi/pkg/utils"
)

func transformLogsToData(logs []line) ([]byte, error) {
	data := make([]byte, 0)
	w := bytes.NewBuffer(data)
	csvW := csv.NewWriter(w)
	err := utils.Apply(logs, func(l line) error {
		wErr := csvW.Write(l.getFieldsAsStringArray())
		if wErr != nil {
			err := errors.Wrap(wErr, "erreur lors de la transformation de la ligne de log "+l.String())
			return err
		}
		return nil
	})
	return data, err
}
