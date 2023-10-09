package stats

import (
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
	slog.Info("fichier de sortie du test", slog.String("filename", f.Name()))
	nbActivities := 100
	knownActivite := createFakeActivite()
	activitesParUtilisateur := createFakeActivites(nbActivities, knownActivite)

	// WHEN
	err = writeToExcel(f, activitesParUtilisateur)
	require.NoError(t, err)

	// THEN
	excel, err := excelize.OpenFile(f.Name())
	require.NoError(t, err)
	firstSheet := excel.WorkBook.Sheets.Sheet[0].Name
	rows, err := excel.GetRows(firstSheet)
	require.NoError(t, err)
	assert.Equal(t, nbActivities, len(rows))
	assert.Equal(t, knownActivite.username, rows[0][0])
	assert.Equal(t, knownActivite.visites, rows[0][1])
	assert.Equal(t, knownActivite.actions, rows[0][2])
	assert.Equal(t, knownActivite.segment, rows[0][3])
	for _, row := range rows {
		for _, colCell := range row {
			fmt.Print(colCell, "\t")
		}
		fmt.Println()
	}
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
