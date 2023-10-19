package stats

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

var fakeTU faker.Faker
var segments []string
var usernames []string

func init() {
	fakeTU = faker.New()
	segments = []string{"bdf", "ddfip", "urssaf", "sf"}
	usernames = []string{
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
		fakeTU.Internet().Email(),
	}
}

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

func TestLog_writeActiviteJourToExcel(t *testing.T) {
	// GIVEN
	f, err := os.CreateTemp(os.TempDir(), "*.xlsx")
	require.NoError(t, err)
	slog.Debug("fichier de sortie du test", slog.String("filename", f.Name()))
	nbActivities := 100
	firstActivite, activitesParJour := createFakeActivitesChan(nbActivities, createFakeActiviteJour)

	// WHEN
	xls := newExcel()
	err = writeActivitesJoursToExcel(xls, activitesParJour)
	require.NoError(t, err)
	err = exportTo(f, xls)
	require.NoError(t, err)

	// THEN
	excel, err := excelize.OpenFile(f.Name())
	require.NoError(t, err)

	// on vérifie la présence de l'onglet qui va bien
	assert.Len(t, excel.WorkBook.Sheets.Sheet, 1)
	firstSheetName := excel.WorkBook.Sheets.Sheet[0].Name
	assert.Equal(t, "Activité par jour", firstSheetName)

	rows, err := excel.GetRows(firstSheetName)
	require.NoError(t, err)
	assert.Len(t, rows, nbActivities)
	assert.Len(t, rows[0], 6)
	assert.Equal(t, firstActivite.jour.Format(time.DateOnly), rows[0][0])
	assert.Equal(t, firstActivite.username, rows[0][1])
	assert.Equal(t, strconv.Itoa(firstActivite.actions), rows[0][2])
	assert.Equal(t, strconv.Itoa(firstActivite.recherches), rows[0][3])
	assert.Equal(t, strconv.Itoa(firstActivite.fiches), rows[0][4])
	assert.Equal(t, firstActivite.segment, rows[0][5])
	for _, row := range rows {
		for _, colCell := range row {
			fmt.Print(colCell, "\t")
		}
		fmt.Println()
	}
}

func createFakeActivitesChan[A any](activitiesNumber int, newActivite func() A) (A, chan row[A]) {
	r := make(chan row[A])
	activite := newActivite()
	go func() {
		defer close(r)
		r <- newRow(activite)
		for i := 1; i < activitiesNumber; i++ {
			r <- newRow(newActivite())
		}
	}()
	return activite, r
}

func createFakeActiviteJour() activiteParJour {
	r := fakeTU.IntBetween(1, 99)
	f := fakeTU.IntBetween(3, 99)
	a := fakeTU.IntBetween(3, 99) + r + f
	return activiteParJour{
		jour:       fakeTU.Time().TimeBetween(time.Now().AddDate(0, -3, 0), time.Now()),
		username:   fakeTU.RandomStringElement(usernames),
		actions:    a,
		recherches: r,
		fiches:     f,
		segment:    fakeTU.RandomStringElement(segments),
	}
}

func createFakeActiviteUtilisateur() activiteParUtilisateur {
	var visites, actions int
	visites = fakeTU.IntBetween(1, 99)
	actions = visites * fakeTU.IntBetween(1, 11)
	return activiteParUtilisateur{
		username: fakeTU.Internet().Email(),
		actions:  strconv.Itoa(actions),
		visites:  strconv.Itoa(visites),
		segment:  fakeTU.RandomStringElement(segments),
	}
}
