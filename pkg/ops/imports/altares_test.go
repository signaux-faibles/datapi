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

func Test_PaydexReader(t *testing.T) {
	ass := assert.New(t)

	// GIVEN

	// workaround: le répertoire de travail change lorsque le tags integration est sélectionné
	path := getTestDataFilename()

	paydexChan, err := PaydexReader(context.Background(), path)
	ass.NoError(err)

	// WHEN
	var paydexes []Paydex
	for current := range paydexChan {
		ass.NoError(current.err)
		paydexes = append(paydexes, current)
	}

	expectedJSON, _ := json.MarshalIndent(expected, " ", " ")
	paydexxJSON, _ := json.MarshalIndent(paydexes, " ", " ")

	// THEN
	ass.Equal(string(expectedJSON), string(paydexxJSON))
}

func Test_PaydexReader_cancelledContext(t *testing.T) {
	ass := assert.New(t)

	// GIVEN
	// workaround: le répertoire de travail change lorsque le tags integration est sélectionné
	path := getTestDataFilename()

	ctx, cancelCtx := context.WithCancel(context.Background())
	paydexChan, _ := PaydexReader(ctx, path)
	cancelCtx()
	var paydexes []Paydex
	for current := range paydexChan {
		paydexes = append(paydexes, current)
	}

	// THEN
	ass.Less(len(paydexes), 4)
}

func Test_PaydexReader_missingFile(t *testing.T) {
	ass := assert.New(t)

	// GIVEN
	// workaround: le répertoire de travail change lorsque le tags integration est sélectionné
	path := "bad" + getTestDataFilename()
	paydexChan, err := PaydexReader(context.Background(), path)

	// WHEN
	var paydexes []Paydex
	for current := range paydexChan {
		paydexes = append(paydexes, current)
	}
	// THEN
	ass.Error(err)
	ass.Len(paydexes, 0)
}

func Test_CopyFromPaydex_firstOccurence(t *testing.T) {
	ass := assert.New(t)

	// GIVEN
	// workaround: le répertoire de travail change lorsque le tags integration est sélectionné
	path := getTestDataFilename()

	paydexReader, _ := PaydexReader(context.Background(), path)
	copyFromPaydex := CopyFromPaydex{
		PaydexParser: paydexReader,
		Current:      new(Paydex),
		Count:        new(int),
	}

	// WHEN
	ok := copyFromPaydex.Next()

	// THEN
	ass.True(ok)
	ass.Equal(1, *copyFromPaydex.Count)
}

func Test_CopyFromPaydex_allTuples(t *testing.T) {
	ass := assert.New(t)

	// GIVEN
	// workaround: le répertoire de travail change lorsque le tags integration est sélectionné
	path := getTestDataFilename()

	paydexReader, _ := PaydexReader(context.Background(), path)
	copyFromPaydex := CopyFromPaydex{
		PaydexParser: paydexReader,
		Current:      new(Paydex),
		Count:        new(int),
	}
	expectedTuples := utils.Convert(expected, Paydex.tuple)

	// WHEN
	var actualTuples [][]interface{}
	for copyFromPaydex.Next() {
		values, _ := copyFromPaydex.Values()
		actualTuples = append(actualTuples, values)
	}

	// THEN
	actualJSON, _ := json.MarshalIndent(actualTuples, " ", " ")
	expectedJSON, _ := json.MarshalIndent(expectedTuples, " ", " ")
	ass.Equal(string(expectedJSON), string(actualJSON))
}

func Test_CopyFromPaydex_missingFile(t *testing.T) {
	ass := assert.New(t)

	// GIVEN
	// workaround: le répertoire de travail change lorsque le tags integration est sélectionné
	path := "bad" + getTestDataFilename()

	paydexReader, _ := PaydexReader(context.Background(), path)
	copyFromPaydex := CopyFromPaydex{
		PaydexParser: paydexReader,
		Current:      new(Paydex),
		Count:        new(int),
	}

	// WHEN
	ok := copyFromPaydex.Next()

	// THEN
	ass.False(ok)
}

func codePaydexForTest(value string) *CodePaydex {
	code, _ := codePaydexFrom(value)
	return code
}
