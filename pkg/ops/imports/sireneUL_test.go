package imports

import (
	"context"
	"datapi/pkg/utils"
	_ "embed"
	"encoding/json"
	"github.com/signaux-faibles/goSirene"
	"github.com/stretchr/testify/assert"
	"testing"
)

//go:embed test_expectedSireneUL.json
var expectedSireneULJSON string

func TestCopyFromSireneUL(t *testing.T) {
	ass := assert.New(t)

	// GIVEN
	var expectedSireneUL []goSirene.SireneUL
	json.Unmarshal([]byte(expectedSireneULJSON), &expectedSireneUL)
	expectedSireneULTuples := utils.Convert(expectedSireneUL, sireneULdata)

	sireneULParser := goSirene.SireneULParser(context.Background(), "test_StockUniteLegale_utf8.zip")

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
