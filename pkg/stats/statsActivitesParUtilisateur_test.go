package stats

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"datapi/pkg/test"
)

func Test_write_writeActivitesUtilisateurToExcel(t *testing.T) {
	xls := newExcel()
	first, items := createFakeActivitesChan(3, createFakeActiviteUtilisateur)

	// écriture dans le fichier

	sheetConf := activitesUtilisateurSheetConfig()
	err := writeOneSheetToExcel(xls, sheetConf, items)
	require.NoError(t, err)

	// sauvegarde dans un fichier
	output, err := os.CreateTemp(os.TempDir(), t.Name()+"_*.xlsx")
	require.NoError(t, err)
	slog.Info("export de l'activité utilisateur", slog.String("filename", output.Name()))
	err = xls.Write(output)
	require.NoError(t, err)

	// vérifications
	rows, err := xls.Rows(sheetConf.label())
	require.NoError(t, err)

	require.True(t, rows.Next()) // il y a au moins la ligne des headers
	headerValues, err := rows.Columns()
	require.NoError(t, err)
	headers, err := sheetConf.headers()
	require.NoError(t, err)
	assert.ElementsMatch(t, headers, headerValues)

	require.True(t, rows.Next())
	firstRowDataValues, err := rows.Columns()
	require.NoError(t, err)
	expected := sheetConf.toRow(first)
	assert.Equal(t, expected[0], firstRowDataValues[0])
	assert.Equal(t, expected[1], test.ToInt(t, firstRowDataValues[1]))
	assert.Equal(t, expected[2], test.ToInt(t, firstRowDataValues[2]))
	assert.Equal(t, expected[3], firstRowDataValues[3])
}
