//go:build integration

package main

import (
	_ "embed"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"datapi/pkg/stats"
	"datapi/pkg/test"
)

func TestAPI_get_stats_since_5_days_ago(t *testing.T) {
	t.Cleanup(func() { test.EraseAccessLogs(t) })
	ass := assert.New(t)
	tuTime := time.Now()

	// GIVEN
	var numberOfDays = 10
	_ = test.InsertSomeLogsAtTime(tuTime.AddDate(0, 0, -numberOfDays-1))
	same := test.InsertSomeLogsAtTime(tuTime.AddDate(0, 0, -numberOfDays))
	after := test.InsertSomeLogsAtTime(tuTime.AddDate(0, 0, -numberOfDays+1))

	path := fmt.Sprintf("/stats/since/%d/days", numberOfDays)
	response := test.HTTPGet(t, path)

	ass.Equal(http.StatusOK, response.StatusCode)
	body := test.GetBodyQuietly(response)
	records, err := test.ReadZippedCSV(body)
	defer test.WriteFile(t.Name()+".zip", body)
	ass.NoError(err)
	// on compare le nombre de lignes lues dans le fichier
	// et le nombre de lignes insérés en base
	ass.Equal(same+after, len(records))
}

func TestAPI_get_stats_with_heavy_load(t *testing.T) {
	t.Cleanup(func() { test.EraseAccessLogs(t) })
	ass := assert.New(t)
	t.Log(t.Name(), "DEMARRAGE")
	t.Log(t.Name(), "debut des insertions")

	// GIVEN
	total := 99999
	var expected int
	min := time.Date(2001, 1, 1, 1, 2, 3, 4, time.UTC)
	max := time.Now()

	for expected = 0; expected < total; {
		t.Log(t.Name(), "il manque des insertions", total-expected)
		expected += test.InsertSomeLogsAtTime(fake.Time().TimeBetween(min, max))
	}
	t.Log(t.Name(), "fin des insertions")

	path := "/stats/from/2000-01-01/to/2099-01-01"
	t.Log(t.Name(), "début de l'appel")
	start := time.Now()
	response := test.HTTPGet(t, path)
	elapsed := time.Since(start)
	t.Log(t.Name(), "fin de l'appel :", elapsed)

	ass.Equal(http.StatusOK, response.StatusCode)
	body := test.GetBodyQuietly(response)
	records, err := test.ReadZippedCSV(body)
	//ass.NoError(err, "contenu de la réponse : %s", body)
	ass.NoError(err)
	t.Logf("nombre de lignes -> %d", len(records))
	ass.Equal(expected, len(records))
	// on compare le nombre de lignes lues dans le fichier
	// et le nombre de lignes insérés en base
}

func TestAPI_get_stats_on_2023_03_01(t *testing.T) {
	t.Cleanup(func() { test.EraseAccessLogs(t) })
	ass := assert.New(t)

	// GIVEN
	search := "2023-03-01"
	nbDays := 3
	dayToSelect, err := time.Parse(stats.DATE_FORMAT, search)
	ass.NoError(err)
	_ = test.InsertSomeLogsAtTime(dayToSelect.Add(-1 * time.Millisecond))
	betweenStart := test.InsertSomeLogsAtTime(dayToSelect)
	betweenEnd := test.InsertSomeLogsAtTime(dayToSelect.AddDate(0, 0, nbDays).Add(-1 * time.Millisecond))
	_ = test.InsertSomeLogsAtTime(dayToSelect.AddDate(0, 0, nbDays+1))
	path := fmt.Sprintf("/stats/from/%s/for/%d/days", search, nbDays)
	response := test.HTTPGet(t, path)

	ass.Equal(http.StatusOK, response.StatusCode)
	body := test.GetBodyQuietly(response)
	defer test.WriteFile(t.Name()+".zip", body)
	records, err := test.ReadZippedCSV(body)
	ass.NoError(err)
	// on compare le nombre de lignes lues dans le fichier
	// et le nombre de lignes insérés en base
	ass.Equal(betweenStart+betweenEnd, len(records))
}

func TestAPI_get_stats_testing_params(t *testing.T) {
	ass := assert.New(t)
	type want struct {
		code    int
		message string
	}
	tests := []struct {
		name string
		path string
		want want
	}{
		{
			"mauvais format de date - 1",
			"/stats/from/20/for/toto/days",
			want{http.StatusBadRequest, "le paramètre 'start' n'est pas du bon type:"},
		},
		{
			"mauvais format de date - 2",
			"/stats/from/20/for/toto/months",
			want{http.StatusBadRequest, "le paramètre 'start' n'est pas du bon type:"},
		},
		{
			"mauvais format de date - 3",
			"/stats/from/2012-03-21/for/any/days",
			want{http.StatusBadRequest, "le paramètre 'n' n'est pas du bon type:"},
		},
		{
			"mauvais format de date - 4",
			"/stats/from/2012-03-21/for/any/months",
			want{http.StatusBadRequest, "le paramètre 'n' n'est pas du bon type:"},
		},
		{
			"mauvais format de date - 5",
			"/stats/since/any/months",
			want{http.StatusBadRequest, "le paramètre 'n' n'est pas du bon type:"},
		},
		{
			"mauvais format de date - 6",
			"/stats/since/any/days",
			want{http.StatusBadRequest, "le paramètre 'n' n'est pas du bon type:"},
		},
		{
			"mauvais format de date - 7",
			"/stats/from/2012-03-21/to/2012-03",
			want{http.StatusBadRequest, "le paramètre 'end' n'est pas du bon type:"},
		},
		{
			"mauvais critère - 1",
			"/stats/from/2012-03-21/to/2012-03-22",
			want{http.StatusBadRequest, "aucune logs trouvée, les critères sont trop restrictifs"},
		},
		{
			"mauvais critère - 2",
			"/stats/since/-1/days",
			want{http.StatusBadRequest, "aucune logs trouvée, les critères sont trop restrictifs"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := test.HTTPGet(t, tt.path)
			body := test.GetBodyQuietly(response)
			t.Log("body --> ", string(body))
			ass.Equal(tt.want.code, response.StatusCode)
			ass.Contains(string(body), tt.want.message)
		})
	}
}
