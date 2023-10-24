package stats

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
)

func Test_write_statsAccessLogsSheet(t *testing.T) {
	xls := newExcel()
	first, items := createFakeActivitesChan(3, createFakeAccessLog)

	// écriture dans le fichier

	sheetConf := accessLogSheetConfig()
	err := writeOneSheetToExcel2(xls, sheetConf, items)
	require.NoError(t, err)

	// sauvegarde dans un fichier
	output, err := os.CreateTemp(os.TempDir(), t.Name()+"_*.xlsx")
	require.NoError(t, err)
	slog.Info("export du access log", slog.String("filename", output.Name()))
	err = xls.Write(output)
	require.NoError(t, err)

	// vérifications
	rows, err := xls.Rows(sheetConf.label())
	require.NoError(t, err)

	require.True(t, rows.Next())
	headerValues, err := rows.Columns()
	require.NoError(t, err)
	assert.ElementsMatch(t, maps.Keys(accessLogHeaders), headerValues)

	i := 1
	for i < sheetConf.startAt() {
		i++
		require.True(t, rows.Next())
	}
	firstRowDataValues, err := rows.Columns()
	require.NoError(t, err)
	assert.ElementsMatch(t, sheetConf.toRow(first), firstRowDataValues)
}

func createFakeAccessLog() accessLog {
	return accessLog{
		date:     fakeTU.Time().TimeBetween(time.Now().AddDate(-3, 0, 0), time.Now()),
		path:     fakeTU.Directory().Directory(fakeTU.IntBetween(1, 3)),
		method:   fakeTU.Internet().HTTPMethod(),
		username: fakeTU.Internet().User(),
		segment:  fakeTU.RandomStringElement([]string{"bdf", "dgfip", "sf"}),
		roles:    fakeTU.Lorem().Words(fakeTU.IntBetween(1, 99)),
	}
}
