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
  "time"

  "github.com/stretchr/testify/assert"

  "datapi/pkg/stats"
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
  records, err := readGZippedCSV(body)
  ass.NoError(err)
  // on compare le nombre de lignes lues dans le fichier
  // et le nombre de lignes insérés en base
  ass.Equal(same+after, len(records))
}

func TestAPI_get_stats_on_2023_03_01(t *testing.T) {
  ass := assert.New(t)

  // GIVEN
  search := "2023-03-01"
  nbDays := 3
  dayToSelect, err := time.Parse(stats.DATE_FORMAT, search)
  ass.NoError(err)
  _ = insertSomeLogsAtTime(dayToSelect.Add(-1 * time.Millisecond))
  betweenStart := insertSomeLogsAtTime(dayToSelect)
  betweenEnd := insertSomeLogsAtTime(dayToSelect.AddDate(0, 0, nbDays))
  _ = insertSomeLogsAtTime(dayToSelect.AddDate(0, 0, nbDays+1))
  path := fmt.Sprintf("/stats/from/%s/for/%d/days", search, nbDays)
  response := test.HTTPGet(t, path)

  ass.Equal(http.StatusOK, response.StatusCode)
  body := test.GetBodyQuietly(response)
  records, err := readGZippedCSV(body)
  ass.NoError(err)
  // on compare le nombre de lignes lues dans le fichier
  // et le nombre de lignes insérés en base
  ass.Equal(betweenStart+betweenEnd, len(records))
}

//func TestAPI_get_stats_with_bad_format_in_param(t *testing.T) {
//	ass := assert.New(t)
//
//	// GIVEN
//	search := "2023-03-01"
//	nbDays := 3
//	dayToSelect, err := time.Parse(stats.DATE_FORMAT, search)
//	ass.NoError(err)
//	_ = insertSomeLogsAtTime(dayToSelect.Add(-1 * time.Millisecond))
//	betweenStart := insertSomeLogsAtTime(dayToSelect)
//	betweenEnd := insertSomeLogsAtTime(dayToSelect.AddDate(0, 0, nbDays))
//	_ = insertSomeLogsAtTime(dayToSelect.AddDate(0, 0, nbDays+1))
//	path := fmt.Sprintf("/stats/from/%s/for/%s/days", search, "COUC")
//	response := test.HTTPGet(t, path)
//
//	ass.Equal(http.StatusOK, response.StatusCode)
//	body := test.GetBodyQuietly(response)
//	records, err := readGZippedCSV(body)
//	ass.NoError(err)
//	// on compare le nombre de lignes lues dans le fichier
//	// et le nombre de lignes insérés en base
//	ass.Equal(betweenStart+betweenEnd, len(records))
//}

func readGZippedCSV(body []byte) ([][]string, error) {
  var data bytes.Buffer
  _, err := data.Write(body)
  if err != nil {
    return nil, err
  }
  gzipReader, err := gzip.NewReader(&data)
  if err != nil {
    return nil, err
  }
  defer gzipReader.Close()
  csvReader := csv.NewReader(gzipReader)
  records, err := csvReader.ReadAll()
  if err != nil {
    return nil, err
  }
  return records, nil
}
