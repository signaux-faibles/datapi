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
	activitesParUtilisateur := createFakeActivites(nbActivities)

	// WHEN
	err = writeToExcel(f, activitesParUtilisateur)
	require.NoError(t, err)

	// THEN
	// toi qui lis ces lignes,
	// va regarder le fichier excel comme un grand
	// et sache que l'automatisation nuit à la responsabilisation
	excel, err := excelize.OpenFile(f.Name())
	require.NoError(t, err)
	firstSheet := excel.WorkBook.Sheets.Sheet[0].Name
	fmt.Printf("'%s' is first sheet of %d sheets.\n", firstSheet, excel.SheetCount)
	rows, err := excel.GetRows(firstSheet)
	require.NoError(t, err)
	assert.Equal(t, nbActivities, len(rows))
	for _, row := range rows {
		for _, colCell := range row {
			fmt.Print(colCell, "\t")
		}
		fmt.Println()
	}
}

func createFakeActivites(activitiesNumber int) chan activiteParUtilisateur {
	r := make(chan activiteParUtilisateur)
	go func() {
		defer close(r)
		for i := 0; i < activitiesNumber; i++ {
			activite := createFakeActivite()
			slog.Info("ajoute une activité", slog.Any("activité", activite), slog.Int("index", i))
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
