package stats

import (
	"time"

	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
)

const accessLogTitle = "Access logs"

func writeOneAccessLogToExcel(f *excelize.File, sheetName string, ligne accessLog, row int) error {
	var i = 0
	i++
	err := writeString(f, sheetName, ligne.date.Format(time.DateTime), i, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture de la date")
	}
	i++
	err = writeString(f, sheetName, ligne.path, i, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du chemin")
	}
	i++
	err = writeString(f, sheetName, ligne.method, i, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du verbe HTTP")
	}
	i++
	err = writeString(f, sheetName, ligne.username, i, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du nom de l'utilisateur")
	}
	i++
	err = writeString(f, sheetName, ligne.segment, i, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du segment")
	}
	i++
	err = writeStrings(f, sheetName, ligne.roles, i, row)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'écriture du nombre de segments")
	}
	return nil
}
