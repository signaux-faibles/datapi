package imports

import (
	"context"
	_ "embed"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/signaux-faibles/goSirene"
	"github.com/stretchr/testify/assert"

	"datapi/pkg/utils"
)

//go:embed tests/test_expectedGeoSirene.json
var expectedGeoSireneJSON string

func TestCopyFromGeoSirene(t *testing.T) {
	ass := assert.New(t)

	// GIVEN
	var expectedGeoSirene []goSirene.GeoSirene
	json.Unmarshal([]byte(expectedGeoSireneJSON), &expectedGeoSirene)
	expectedTuples := utils.Convert(expectedGeoSirene, geoSireneData)
	expectedTuplesJSON, _ := json.MarshalIndent(expectedTuples, " ", " ")

	// workaround: le répertoire de travail change lorsque le tags integration est sélectionné
	cwd, _ := os.Getwd()
	path := "tests/test_StockEtablissement_utf8_geo.csv.gz"
	if !strings.HasSuffix(cwd, "imports") {
		path = "./pkg/ops/imports/" + path
	}

	file, _ := os.Open(path)
	geoSireneParser := goSirene.GeoSireneParser(context.Background(), file)

	// WHEN
	copyFromGeoSirene := CopyFromGeoSirene{
		GeoSireneParser: geoSireneParser,
		Current:         new(goSirene.GeoSirene),
		Count:           new(int),
	}
	var actualTuples [][]interface{}
	for copyFromGeoSirene.Next() {
		actual, _ := copyFromGeoSirene.Values()
		actualTuples = append(actualTuples, actual)
	}
	actualTuplesJSON, _ := json.MarshalIndent(actualTuples, " ", " ")

	// THEN
	ass.Equal(string(expectedTuplesJSON), string(actualTuplesJSON))
}

// TODO: fix close file on error
func TestCopyFromGeoSirene_missing_file(t *testing.T) {
	ass := assert.New(t)

	// GIVEN
	file, _ := os.Open("missing_test_StockEtablissement_utf8_geo.csv.gz")
	geoSireneParser := goSirene.GeoSireneParser(context.Background(), file)

	// WHEN
	copyFromGeoSirene := CopyFromGeoSirene{
		GeoSireneParser: geoSireneParser,
		Current:         new(goSirene.GeoSirene),
		Count:           new(int),
	}
	ok := copyFromGeoSirene.Next()

	// THEN
	ass.False(ok)
}
