package stats

import (
	"log/slog"
	"os"
	"strconv"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/require"
)

func TestLog_writeToExcel(t *testing.T) {
	// GIVEN
	f, err := os.Create(os.TempDir() + "enveloppe.xlsx")
	require.NoError(t, err)
	slog.Info("fichier de sortie du test", slog.String("filename", f.Name()))
	activitesParUtilisateur := createFakeActivites()

	// WHEN
	err = writeToExcel(f, activitesParUtilisateur)
	require.NoError(t, err)

	// THEN
	// va regarder le fichier excel comme un grand
	// l'automatisation nuit à la responsabilisation
}

func createFakeActivites() chan activiteParUtilisateur {
	r := make(chan activiteParUtilisateur)
	go func() {
		defer close(r)
		for i := 0; i < 99; i++ {
			r <- createFakeActivite()
		}
	}()
	return r
}

func createFakeActivite() activiteParUtilisateur {
	fake := faker.New()
	var visites, actions int
	visites = fake.IntBetween(1, 99)
	actions = visites * fake.IntBetween(1, 11)
	return activiteParUtilisateur{
		username: fake.Internet().Email(),
		actions:  strconv.Itoa(actions),
		visites:  strconv.Itoa(visites),
		segment:  fake.App().Name(),
	}
}
