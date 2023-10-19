package stats

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestLog_writeActivitesUtilisateurToExcel(t *testing.T) {
	// GIVEN
	f, err := os.CreateTemp(os.TempDir(), "*.xlsx")
	require.NoError(t, err)
	slog.Debug("fichier de sortie du test", slog.String("filename", f.Name()))
	nbActivities := 3
	lastActivite, activitesParUtilisateur := createFakeActivitesChan(nbActivities, createFakeActiviteUtilisateur)

	// WHEN
	xls := newExcel()
	err = writeActivitesUtilisateurToExcel(xls, activitesParUtilisateur)
	require.NoError(t, err)
	err = exportTo(f, xls)
	require.NoError(t, err)

	// THEN
	excel, err := excelize.OpenFile(f.Name())
	require.NoError(t, err)

	// on vérifie la présence de l'onglet qui va bien
	assert.Len(t, excel.WorkBook.Sheets.Sheet, 1)
	firstSheetName := excel.WorkBook.Sheets.Sheet[0].Name
	assert.Equal(t, "Activité par utilisateur", firstSheetName)

	rows, err := excel.GetRows(firstSheetName)
	require.NoError(t, err)
	assert.Len(t, rows, nbActivities)
	assert.Len(t, rows[0], 4)
	assert.Equal(t, lastActivite.username, rows[0][0])
	assert.Equal(t, lastActivite.visites, rows[0][1])
	assert.Equal(t, lastActivite.actions, rows[0][2])
	assert.Equal(t, lastActivite.segment, rows[0][3])
	assert.Equal(t, lastActivite.segment, rows[0][3])
	for _, row := range rows {
		for _, colCell := range row {
			fmt.Print(colCell, "\t")
		}
		fmt.Println()
	}
}

func writeActivitesUtilisateurToExcel(xls *excelize.File, activites chan row[activiteParUtilisateur]) error {
	err := writeOneSheetToExcel(xls, "Activité par utilisateur", activites, writeOneActiviteUtilisateurToExcel)
	if err != nil {
		return fmt.Errorf("erreur lors de l'écriture d'une ligne d'activités par utilisateurs : %w", err)
	}
	return nil
}
