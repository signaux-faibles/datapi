//go:build integration

package imports

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"datapi/pkg/test"
)

var tuTime = time.Date(2023, 03, 10, 17, 41, 58, 651387237, time.UTC)

func TestMain(m *testing.M) {
	testConfig := map[string]string{}
	url := test.GetDatapiDBURL()
	testConfig["postgres"] = url
	testConfig["sourceEntreprise"] = test.ProjectPathOf("test/emptyJSON.gz")
	testConfig["sourceEtablissement"] = test.ProjectPathOf("test/emptyJSON.gz")
	if err := test.Viperize(testConfig); err != nil {
		slog.Error("erreur pendant le chargement de la configuration ", slog.Any("error", err))
	}

	// Run tests
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	// on peut placer ici du code de nettoyage si nécessaire

	os.Exit(code)
}

func Test_importEntreprisesAndEtablissement(t *testing.T) {
	test.FakeTime(t, tuTime)
	//ass := assert.New(t)
	err := test.Viperize(nil)
	require.NoError(t, err)
	//net.ParseIP()
	err = importEtablissement()
	require.NoError(t, err)
}
