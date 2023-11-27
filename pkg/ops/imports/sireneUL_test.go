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

//go:embed tests/test_expectedSireneUL.json
var expectedSireneULJSON string

func TestCopyFromSireneUL(t *testing.T) {
	ass := assert.New(t)

	// GIVEN
	var expectedSireneUL []goSirene.SireneUL
	json.Unmarshal([]byte(expectedSireneULJSON), &expectedSireneUL)
	expectedSireneULTuples := utils.Convert(expectedSireneUL, sireneULdata)

	// workaround: le répertoire de travail change lorsque le tags integration est sélectionné
	cwd, _ := os.Getwd()
	path := "tests/test_StockUniteLegale_utf8.zip"
	if !strings.HasSuffix(cwd, "imports") {
		path = "./pkg/ops/imports/" + path
	}

	sireneULParser := goSirene.SireneULParser(context.Background(), path)

	// WHEN
	copyFromSireneUL := CopyFromSireneUL{
		sireneULParser,
		new(goSirene.SireneUL),
		new(int),
	}

	var actualTuples [][]interface{}
	for copyFromSireneUL.Next() {
		actual, _ := copyFromSireneUL.Values()
		actualTuples = append(actualTuples, actual)
	}

	expectedTuplesJSON, _ := json.MarshalIndent(expectedSireneULTuples, " ", " ")
	actualTuplesJSON, _ := json.MarshalIndent(actualTuples, " ", " ")

	// THEN
	ass.Equal(string(expectedTuplesJSON), string(actualTuplesJSON))
}
