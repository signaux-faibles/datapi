package stats

import (
	"archive/zip"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

var fake2 faker.Faker
var segments []string

func init() {
	fake2 = faker.New()
	segments = []string{"bdf", "ddfip", "urssaf", "sf"}
}

func TestLog_writeToExcel(t *testing.T) {
	// GIVEN
	f, err := os.CreateTemp(os.TempDir(), "enveloppe_*.xlsx")
	require.NoError(t, err)
	slog.Debug("fichier de sortie du test", slog.String("filename", f.Name()))
	nbActivities := 100
	knownActivite := createFakeActivite()
	activitesParUtilisateur := createFakeActivites(nbActivities, knownActivite)

	// WHEN
	err = writeToExcel(f, activitesParUtilisateur)
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
	assert.Equal(t, knownActivite.username, rows[0][0])
	assert.Equal(t, knownActivite.visites, rows[0][1])
	assert.Equal(t, knownActivite.actions, rows[0][2])
	assert.Equal(t, knownActivite.segment, rows[0][3])
	assert.Equal(t, knownActivite.segment, rows[0][3])
	for _, row := range rows {
		for _, colCell := range row {
			fmt.Print(colCell, "\t")
		}
		fmt.Println()
	}
}

func TestLog_createExcel(t *testing.T) {
	// GIVEN
	f, err := os.CreateTemp(os.TempDir(), t.Name()+"_*.zip")
	require.NoError(t, err)
	archive := zip.NewWriter(f)
	slog.Info("fichier de sortie du test", slog.String("filename", f.Name()))

	// WHEN
	err = createExcel(archive, t.Name(), nil)
	require.NoError(t, err)
	err = archive.Flush()
	require.NoError(t, err)
	err = archive.Close()
	require.NoError(t, err)

	// THEN
	zip, err := zip.OpenReader(f.Name())
	require.NoError(t, err)
	zipEntry, err := zip.Open(t.Name() + ".xlsx")
	require.NoError(t, err)
	_, err = excelize.OpenReader(zipEntry)
	require.NoError(t, err)
}

func createFakeActivites(activitiesNumber int, activites ...activiteParUtilisateur) chan activiteParUtilisateur {
	r := make(chan activiteParUtilisateur)
	go func() {
		defer close(r)
		for _, activite := range activites {
			r <- activite
		}
		for i := len(activites); i < activitiesNumber; i++ {
			activite := createFakeActivite()
			r <- activite
		}
	}()
	return r
}

func createFakeActivite() activiteParUtilisateur {
	var visites, actions int
	visites = fake2.IntBetween(1, 99)
	actions = visites * fake2.IntBetween(1, 11)
	return activiteParUtilisateur{
		username: fake2.Internet().Email(),
		actions:  strconv.Itoa(actions),
		visites:  strconv.Itoa(visites),
		segment:  fake2.RandomStringElement(segments),
	}
}
