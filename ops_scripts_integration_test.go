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
	"datapi/pkg/ops/scripts"
	"datapi/pkg/test"
	"datapi/pkg/utils"
)

func TestFetchNonExistentScript(t *testing.T) {
	ass := assert.New(t)
	result, err := scripts.Fetch(uuid.New())
	ass.Equal(scripts.Empty, result)
	ass.NotNil(err)
}

func TestRunScript(t *testing.T) {
	ass := assert.New(t)
	current := scripts.StartRefreshScript(context.Background(), db.Get(), scripts.Wait5Seconds)
	t.Logf("refreshID is running with id : %s", current.UUID)
	result, err := scripts.Fetch(current.UUID)
	ass.Nil(err)
	ass.NotNil(result)
}

func TestLastScriptState(t *testing.T) {
	ass := assert.New(t)
	lastRefreshState, err := scripts.FetchLast()
	require.NoError(t, err)
	if lastRefreshState == scripts.Empty {
		scripts.StartRefreshScript(context.Background(), db.Get(), scripts.Wait5Seconds)
	}
	time.Sleep(100 * time.Millisecond)
	lastRefreshState, err = scripts.FetchLast()
	require.NoError(t, err)
	ass.NotEmpty(lastRefreshState)
	t.Logf("Description du dernier refresh : %s", lastRefreshState)
}

func TestFetchRefreshWithState(t *testing.T) {
	ass := assert.New(t)
	failed1 := scripts.StartRefreshScript(context.Background(), db.Get(), scripts.Fail)
	failed2 := scripts.StartRefreshScript(context.Background(), db.Get(), scripts.Fail)

	running := scripts.StartRefreshScript(context.Background(), db.Get(), scripts.Wait5Seconds)
	time.Sleep(1 * time.Second)

	failedRefresh := scripts.FetchRefreshsWithState(scripts.Failed)
	ass.Contains(failedRefresh, *failed1)
	ass.Contains(failedRefresh, *failed2)

	runningRefresh := scripts.FetchRefreshsWithState(scripts.Running)
	getUUID := func(r scripts.Refresh) uuid.UUID { return r.UUID }
	ass.Contains(utils.Convert(runningRefresh, getUUID), running.UUID)
}

func TestApi_OpsScripts_StatusHandler(t *testing.T) {
	ass := assert.New(t)
	current := scripts.StartRefreshScript(context.Background(), db.Get(), scripts.Wait5Seconds)
	time.Sleep(100 * time.Millisecond)
	path := "/ops/scripts/status/" + current.UUID.String()
	response := test.HTTPGet(t, path)

	body := test.GetBodyQuietly(response)
	ass.Equalf(http.StatusOK, response.StatusCode, "body de la réponse : %s", body)

	retour := &scripts.Refresh{}
	err := json.Unmarshal(body, &retour)
	require.NoError(t, err)
	t.Log(retour)
	ass.Equal(current.UUID, retour.UUID)
	ass.Equal(scripts.Running, retour.Status)
}

func TestApi_Ops_ListStatus(t *testing.T) {
	utils.ConfigureLogLevel("warn")
	ass := assert.New(t)
	ctx := context.Background()

	current := scripts.StartRefreshScript(ctx, db.Get(), scripts.Wait5Seconds)
	path := "/ops/scripts/status/" + current.UUID.String()

	// Création d'un canal pour gérer le timeout
	timeout := time.After(6 * time.Second) // Timeout de 5 secondes
	actual := &scripts.Refresh{}
	// Boucle while avec timeout
	for actual.Status != scripts.Finished {
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
	ass.Equal(scripts.Finished, actual.Status)
}

func TestApi_OpsScripts_ListHandler(t *testing.T) {
	ass := assert.New(t)
	test.FakeTime(t, tuTime)

	current := scripts.StartRefreshScript(context.Background(), db.Get(), scripts.Wait5Seconds)
	expectedStatus := scripts.Prepare
	rPath := "/ops/scripts/list/" + string(expectedStatus)

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

func decodeRefreshArray(body []byte) ([]scripts.Refresh, error) {
	var actual []scripts.Refresh
	err := json.Unmarshal(body, &actual)
	if err != nil {
		return nil, errors.Wrap(err, "erreur pendant l'unmarshalling de la réponse")
	}
	return actual, nil
}
