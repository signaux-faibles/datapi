//go:build integration

package main

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/csv"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"datapi/pkg/test"
)

func TestAPI_get_stats_since_5_days_ago(t *testing.T) {
	ass := assert.New(t)

	// GIVEN
	var numberOfDays = 10
	_ = insertSomeLogsAtTime(tuTime.AddDate(0, 0, -numberOfDays-1))
	same := insertSomeLogsAtTime(tuTime.AddDate(0, 0, -numberOfDays))
	after := insertSomeLogsAtTime(tuTime.AddDate(0, 0, -numberOfDays+1))

	path := fmt.Sprintf("/stats/since/%d/days", numberOfDays)
	response := test.HTTPGet(t, path)

	ass.Equal(http.StatusOK, response.StatusCode)
	body := test.GetBodyQuietly(response)
	var data bytes.Buffer
	_, err := data.Write(body)
	ass.NoError(err)
	gzipReader, err := gzip.NewReader(&data)
	ass.NoError(err)
	defer gzipReader.Close()
	csvReader := csv.NewReader(gzipReader)
	records, err := csvReader.ReadAll()
	ass.NoError(err)
	// on compare le nombre de lignes lues dans le fichier
	// et le nombre de lignes insérés en base
	ass.Equal(same+after, len(records))
}
