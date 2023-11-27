package imports

import (
	"context"
	"encoding/json"

	"github.com/stretchr/testify/assert"

	"datapi/pkg/utils"

	"os"
	"strings"
	"testing"
	"time"
)

var expected = []Paydex{
	{
		Siren:        "000000000",
		Paydex:       codePaydexForTest("078"),
		JoursRetard:  pointer(3),
		Fournisseurs: pointer(20),
		Encours:      pointer(float64(104593.746)),
		DateValeur:   pointer(time.Date(2023, 01, 15, 0, 0, 0, 0, time.UTC)),
	},
	{
		Siren:        "000000000",
		Paydex:       codePaydexForTest("078"),
		JoursRetard:  pointer(3),
		Fournisseurs: pointer(21),
		Encours:      pointer(float64(139321.632)),
		DateValeur:   pointer(time.Date(2022, 12, 15, 0, 0, 0, 0, time.UTC)),
	},
	{
		Siren:        "111111111",
		Paydex:       codePaydexForTest("078"),
		JoursRetard:  pointer(3),
		Fournisseurs: pointer(22),
		Encours:      pointer(float64(150023.553)),
		FPI30:        pointer(11),
		FPI90:        pointer(11),
		DateValeur:   pointer(time.Date(2022, 11, 15, 0, 0, 0, 0, time.UTC)),
	},
	{
		Siren:        "111111111",
		Paydex:       codePaydexForTest("078"),
		JoursRetard:  pointer(3),
		Fournisseurs: pointer(22),
		Encours:      pointer(float64(150785.798)),
		FPI30:        pointer(11),
		FPI90:        pointer(11),
		DateValeur:   pointer(time.Date(2022, 10, 15, 0, 0, 0, 0, time.UTC)),
	},
}

func getTestDataFilename() string {
	cwd, _ := os.Getwd()
	path := "tests/test_paydex.zip"
	if !strings.HasSuffix(cwd, "imports") {
		path = "./pkg/ops/imports/" + path
	}
	return path
}

func Test_PaydexHistoReader(t *testing.T) {
	ass := assert.New(t)

	// GIVEN

	// workaround: le répertoire de travail change lorsque le tags integration est sélectionné
	path := getTestDataFilename()

	paydexHistoChan, err := PaydexReader(context.Background(), path)
	ass.NoError(err)

	// WHEN
	var paydexHistos []Paydex
	for paydexHisto := range paydexHistoChan {
		ass.NoError(paydexHisto.err)
		paydexHistos = append(paydexHistos, paydexHisto)
	}

	expectedJSON, _ := json.MarshalIndent(expected, " ", " ")
	paydexHistosJSON, _ := json.MarshalIndent(paydexHistos, " ", " ")

	// THEN
	ass.Equal(string(expectedJSON), string(paydexHistosJSON))
}

func Test_PaydexHistoReader_cancelledContext(t *testing.T) {
	ass := assert.New(t)

	// GIVEN
	// workaround: le répertoire de travail change lorsque le tags integration est sélectionné
	path := getTestDataFilename()

	ctx, cancelCtx := context.WithCancel(context.Background())
	paydexHistoChan, _ := PaydexReader(ctx, path)
	cancelCtx()
	var paydexHistos []Paydex
	for paydexHisto := range paydexHistoChan {
		paydexHistos = append(paydexHistos, paydexHisto)
	}

	// THEN
	ass.Less(len(paydexHistos), 4)
}

func Test_PaydexHistoReader_missingFile(t *testing.T) {
	ass := assert.New(t)

	// GIVEN
	// workaround: le répertoire de travail change lorsque le tags integration est sélectionné
	path := "bad" + getTestDataFilename()
	paydexHistoChan, err := PaydexReader(context.Background(), path)

	// WHEN
	var paydexHistos []Paydex
	for paydexHisto := range paydexHistoChan {
		paydexHistos = append(paydexHistos, paydexHisto)
	}
	// THEN
	ass.Error(err)
	ass.Len(paydexHistos, 0)
}

func Test_CopyFromPaydexHisto_firstOccurence(t *testing.T) {
	ass := assert.New(t)

	// GIVEN
	// workaround: le répertoire de travail change lorsque le tags integration est sélectionné
	path := getTestDataFilename()

	paydexHistoReader, _ := PaydexReader(context.Background(), path)
	copyFromPaydexHisto := CopyFromPaydex{
		PaydexParser: paydexHistoReader,
		Current:      new(Paydex),
		Count:        new(int),
	}

	// WHEN
	ok := copyFromPaydexHisto.Next()

	// THEN
	ass.True(ok)
	ass.Equal(1, *copyFromPaydexHisto.Count)
}

func Test_CopyFromPaydexHisto_allTuples(t *testing.T) {
	ass := assert.New(t)

	// GIVEN
	// workaround: le répertoire de travail change lorsque le tags integration est sélectionné
	path := getTestDataFilename()

	paydexHistoReader, _ := PaydexReader(context.Background(), path)
	copyFromPaydexHisto := CopyFromPaydex{
		PaydexParser: paydexHistoReader,
		Current:      new(Paydex),
		Count:        new(int),
	}
	expectedTuples := utils.Convert(expected, Paydex.tuple)

	// WHEN
	var actualTuples [][]interface{}
	for copyFromPaydexHisto.Next() {
		values, _ := copyFromPaydexHisto.Values()
		actualTuples = append(actualTuples, values)
	}

	// THEN
	actualJSON, _ := json.MarshalIndent(actualTuples, " ", " ")
	expectedJSON, _ := json.MarshalIndent(expectedTuples, " ", " ")
	ass.Equal(string(expectedJSON), string(actualJSON))
}

func Test_CopyFromPaydexHisto_missingFile(t *testing.T) {
	ass := assert.New(t)

	// GIVEN
	// workaround: le répertoire de travail change lorsque le tags integration est sélectionné
	path := "bad" + getTestDataFilename()

	paydexHistoReader, _ := PaydexReader(context.Background(), path)
	copyFromPaydexHisto := CopyFromPaydex{
		PaydexParser: paydexHistoReader,
		Current:      new(Paydex),
		Count:        new(int),
	}

	// WHEN
	ok := copyFromPaydexHisto.Next()

	// THEN
	ass.False(ok)
}

func codePaydexForTest(value string) *CodePaydex {
	code, _ := codePaydexFrom(value)
	return code
}
