//go:build integration

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"datapi/pkg/db"
	"datapi/pkg/ops/refresh"
	"datapi/pkg/test"
	"datapi/pkg/utils"
)

func TestFetchNonExistentRefresh(t *testing.T) {
	ass := assert.New(t)
	result, err := refresh.Fetch(uuid.New())
	ass.Equal(refresh.Empty, result)
	ass.NotNil(err)
}

func TestRunRefreshScript(t *testing.T) {
	ass := assert.New(t)
	current := refresh.StartRefreshScript(context.Background(), db.Get(), refresh.SQLPopulateVTables)
	t.Logf("refreshID is running with id : %s", current.UUID)
	result, err := refresh.Fetch(current.UUID)
	ass.Nil(err)
	ass.NotNil(result)
}

func TestLastRefreshState(t *testing.T) {
	ass := assert.New(t)
	lastRefreshState, err := refresh.FetchLast()
	require.NoError(t, err)
	if lastRefreshState == refresh.Empty {
		refresh.StartRefreshScript(context.Background(), db.Get(), refresh.SQLPopulateVTables)
	}
	time.Sleep(100 * time.Millisecond)
	lastRefreshState, err = refresh.FetchLast()
	require.NoError(t, err)
	ass.NotEmpty(lastRefreshState)
	t.Logf("Description du dernier refresh : %s", lastRefreshState)
}

func TestApi_OpsRefresh_StartHandler(t *testing.T) {
	ass := assert.New(t)
	path := "/ops/refresh/vtables"
	response := test.HTTPGet(t, path)

	ass.Equalf(http.StatusOK, response.StatusCode, "body de la réponse : %s", test.GetBodyQuietly(response))
}

func TestFetchRefreshWithState(t *testing.T) {
	ass := assert.New(t)
	failed1 := refresh.StartRefreshScript(context.Background(), db.Get(), "sql incorrect")
	failed2 := refresh.StartRefreshScript(context.Background(), db.Get(), "sql incorrect")

	running := refresh.StartRefreshScript(context.Background(), db.Get(), "SELECT pg_sleep(5);")
	time.Sleep(1 * time.Second)

	failedRefresh := refresh.FetchRefreshsWithState(refresh.Failed)
	ass.Contains(failedRefresh, *failed1)
	ass.Contains(failedRefresh, *failed2)

	runningRefresh := refresh.FetchRefreshsWithState(refresh.Running)
	getUUID := func(r refresh.Refresh) uuid.UUID { return r.UUID }
	ass.Contains(utils.Convert(runningRefresh, getUUID), running.UUID)
}

func TestApi_OpsRefresh_StatusHandler(t *testing.T) {
	ass := assert.New(t)
	rollbackCtx, rollback := context.WithCancel(context.Background())
	current := refresh.StartRefreshScript(rollbackCtx, db.Get(), "SELECT CURRENT_DATE")
	path := "/ops/refresh/status/" + current.UUID.String()
	response := test.HTTPGet(t, path)
	rollback()

	body := test.GetBodyQuietly(response)
	ass.Equalf(http.StatusOK, response.StatusCode, "body de la réponse : %s", body)

	retour := &refresh.Refresh{}
	err := json.Unmarshal(body, &retour)
	require.NoError(t, err)
	t.Log(retour)
	ass.Equal(current.UUID, retour.UUID)
	ass.Equal(current.Status, refresh.Failed)
}

func TestApi_Ops_RefreshVTables(t *testing.T) {
	utils.ConfigureLogLevel("warn")
	ass := assert.New(t)
	ctx := context.Background()

	current := refresh.StartRefreshScript(ctx, db.Get(), refresh.SQLPopulateVTables)
	path := "/ops/refresh/status/" + current.UUID.String()

	// Création d'un canal pour gérer le timeout
	timeout := time.After(10 * time.Second) // Timeout de 5 secondes
	actual := &refresh.Refresh{}
	// Boucle while avec timeout
	for actual.Status != refresh.Finished && actual.Status != refresh.Failed {
		time.Sleep(1 * time.Second)
		select {
		case <-timeout:
			t.Error("time out !!!")
			t.Fail()
			return
		default:
			response := test.HTTPGet(t, path)
			body := test.GetBodyQuietly(response)
			ass.Equalf(http.StatusOK, response.StatusCode, "body de la réponse : %s", body)

			err := json.Unmarshal(body, &actual)
			t.Logf("refresh : %s", actual)
			require.NoError(t, err)
		}
	}

	ass.Equal(current.UUID, actual.UUID)
	ass.Equal(refresh.Finished, actual.Status)
}

func TestApi_OpsRefresh_ListHandler(t *testing.T) {
	ass := assert.New(t)
	test.FakeTime(t, tuTime)

	current := refresh.StartRefreshScript(context.Background(), db.Get(), "SELECT CURRENT_DATE")
	expectedStatus := refresh.Prepare
	rPath := "/ops/refresh/list/" + string(expectedStatus)

	response := test.HTTPGet(t, rPath)
	body := test.GetBodyQuietly(response)
	ass.Equal(http.StatusOK, response.StatusCode, "body de la réponse : %s", body)
	actual, err := decodeRefreshArray(body)
	ass.Nil(err)
	ass.Contains(actual, *current)
	ass.Conditionf(func() bool {
		// tous les refresh doivent avoir le status `failed`
		for _, current := range actual {
			if current.Status != expectedStatus {
				return false
			}
		}
		return true
	}, "Un des refresh n'a pas le status `failed`")
}

func decodeRefreshArray(body []byte) ([]refresh.Refresh, error) {
	var actual []refresh.Refresh
	err := json.Unmarshal(body, &actual)
	if err != nil {
		return nil, errors.Wrap(err, "erreur pendant l'unmarshalling de la réponse")
	}
	return actual, nil
}
