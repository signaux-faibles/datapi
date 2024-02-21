//go:build integration

package imports

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"datapi/pkg/db"
	"datapi/pkg/test"
)

var tuTime = time.Date(2023, 03, 10, 17, 41, 58, 651387237, time.UTC)
var fake = test.NewFaker()

func TestMain(m *testing.M) {
	testConfig := map[string]string{}
	url := test.GetDatapiDBURL()
	testConfig["postgres"] = url
	testConfig["sourceEntreprise"] = test.ProjectPathOf("test/emptyJSON.gz")
	testConfig["sourceEtablissement"] = test.ProjectPathOf("test/emptyJSON.gz")
	if err := test.Viperize(testConfig); err != nil {
		slog.Error("erreur pendant le chargement de la configuration ", slog.Any("error", err))
	}
	insertEtablissement4Predictions()
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

func Test_importPredictions_then_deletePredictions(t *testing.T) {
	// GIVEN
	batchNumber := "2304"
	algo := "tu"
	listPath := test.ProjectPathOf("test/sampleAvec5Predictions.json")
	config := map[string]string{}
	config["source.listPath"] = listPath
	err := test.Viperize(config)
	require.NoError(t, err)

	// WHEN
	err = importPredictions(batchNumber, algo)
	require.NoError(t, err)
	err = importPredictions(batchNumber, fake.Lorem().Word())
	require.NoError(t, err)
	err = importPredictions(strconv.Itoa(fake.IntBetween(1000, 9999)), algo)
	require.NoError(t, err)
	err = importPredictions(strconv.Itoa(fake.IntBetween(1000, 9999)), fake.Lorem().Word())
	require.NoError(t, err)

	lignesSupprimees, err := deletePredictions(batchNumber, algo)
	require.NoError(t, err)

	// THEN (seules les scores des entreprises insérées avec batchNumber et algo doivent être supprimées
	assert.Equal(t, int64(5), lignesSupprimees)
}

func insertEtablissement4Predictions() {
	filePath := test.ProjectPathOf("test/insere5EtablissementsPourLesPredictions.sql")
	content, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	_, err = db.Get().Exec(context.Background(), string(content))
	if err != nil {
		panic(err)
	}
}
