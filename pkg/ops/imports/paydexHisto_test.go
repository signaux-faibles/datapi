package imports

import (
	"context"
	"datapi/pkg/utils"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
	"time"
)

var expected = []PaydexHisto{
	{
		Siren:        "000000000",
		Paydex:       pointer("035"),
		JoursRetard:  pointer(75),
		Fournisseurs: pointer(16),
		Encours:      pointer(float64(93131)),
		DateValeur:   pointer(time.Date(2015, 02, 15, 0, 0, 0, 0, time.UTC)),
	},
	{
		Siren:        "000000000",
		Paydex:       pointer("034"),
		JoursRetard:  pointer(78),
		Fournisseurs: pointer(17),
		Encours:      pointer(float64(109474)),
		DateValeur:   pointer(time.Date(2015, 01, 15, 0, 0, 0, 0, time.UTC)),
	},
	{
		Siren:               "111111111",
		Paydex:              pointer("079"),
		JoursRetard:         pointer(1),
		Fournisseurs:        pointer(18),
		Encours:             pointer(float64(885692.202)),
		ExperiencesPaiement: pointer(32),
		FPI30:               pointer(0),
		FPI90:               pointer(0),
		DateValeur:          pointer(time.Date(2023, 07, 06, 0, 0, 0, 0, time.UTC)),
	},
	{
		Siren:               "111111111",
		Paydex:              pointer("079"),
		JoursRetard:         pointer(1),
		Fournisseurs:        pointer(18),
		Encours:             pointer(float64(885930.023)),
		ExperiencesPaiement: pointer(33),
		FPI30:               pointer(0),
		FPI90:               pointer(0),
		DateValeur:          pointer(time.Date(2023, 06, 15, 0, 0, 0, 0, time.UTC)),
	},
}

func getTestDataFilename() string {
	cwd, _ := os.Getwd()
	path := "tests/test_histo_paydex.csv.zip"
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

	paydexHistoChan, _ := PaydexHistoReader(context.Background(), path)

	// WHEN
	var paydexHistos []PaydexHisto
	for paydexHisto := range paydexHistoChan {
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
	paydexHistoChan, _ := PaydexHistoReader(ctx, path)
	cancelCtx()
	var paydexHistos []PaydexHisto
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
	paydexHistoChan, err := PaydexHistoReader(context.Background(), path)

	// WHEN
	var paydexHistos []PaydexHisto
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

	paydexHistoReader, _ := PaydexHistoReader(context.Background(), path)
	copyFromPaydexHisto := CopyFromPaydexHisto{
		PaydexHistoParser: paydexHistoReader,
		Current:           new(PaydexHisto),
		Count:             new(int),
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

	paydexHistoReader, _ := PaydexHistoReader(context.Background(), path)
	copyFromPaydexHisto := CopyFromPaydexHisto{
		PaydexHistoParser: paydexHistoReader,
		Current:           new(PaydexHisto),
		Count:             new(int),
	}
	expectedTuples := utils.Convert(expected, PaydexHisto.tuple)

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

	paydexHistoReader, _ := PaydexHistoReader(context.Background(), path)
	copyFromPaydexHisto := CopyFromPaydexHisto{
		PaydexHistoParser: paydexHistoReader,
		Current:           new(PaydexHisto),
		Count:             new(int),
	}

	// WHEN
	ok := copyFromPaydexHisto.Next()

	// THEN
	ass.False(ok)
}
