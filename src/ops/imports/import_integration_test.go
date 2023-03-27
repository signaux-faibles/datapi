//go:build integration
// +build integration

package imports

import (
	"github.com/signaux-faibles/datapi/src/test"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var tuTime = time.Date(2023, 03, 10, 17, 41, 58, 651387237, time.UTC)

func TestMain(m *testing.M) {
	test.FakeTime(tuTime)
	testConfig := initTestConfig()
	url := test.GetDatapiDbURL()
	testConfig["postgres"] = url
	testConfig["sourceEntreprise"] = "test/emptyJSON.gz"
	testConfig["sourceEtablissement"] = "test/emptyJSON.gz"
	if err := test.Viperize(testConfig); err != nil {
		log.Fatalf("erreur pendant le chargement de la configuration ")
	}
	test.Wait4DatapiDb(url)
	//Run tests
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	// on peut placer ici du code de nettoyage si nécessaire

	os.Exit(code)
}

func Test_importEntreprisesAndEtablissement(t *testing.T) {

	ass := assert.New(t)
	//net.ParseIP()
	err := importEntreprisesAndEtablissement()
	ass.Nil(err)

}

func initTestConfig() map[string]string {

	root, _ := os.Getwd()
	viper.AddConfigPath(filepath.Join(root, "test"))
	testConfig := map[string]string{}
	testConfig["migrationsDir"] = filepath.Join(root, "migrations")
	return testConfig
}
