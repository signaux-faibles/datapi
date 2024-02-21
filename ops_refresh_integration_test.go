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
	current := refresh.StartRefreshScript(context.Background(), db.Get())
	t.Logf("refreshID is running with id : %s", current.UUID)
	result, err := refresh.Fetch(current.UUID)
	ass.Nil(err)
	ass.NotNil(result)
}

func TestLastRefreshState(t *testing.T) {
	ass := assert.New(t)
	lastRefreshState := refresh.FetchLast()
	if lastRefreshState == refresh.Empty {
		refresh.StartRefreshScript(context.Background(), db.Get())
	}
	time.Sleep(100 * time.Millisecond)
	lastRefreshState = refresh.FetchLast()
	ass.NotEmpty(lastRefreshState)
	t.Logf("Description du dernier refresh : %s", lastRefreshState)
}

func TestApi_OpsRefresh_StartHandler(t *testing.T) {
	ass := assert.New(t)
	path := "/ops/refresh/start"
	response := test.HTTPGet(t, path)

	ass.Equalf(http.StatusOK, response.StatusCode, "body de la réponse : %s", test.GetBodyQuietly(response))
}

func TestFetchRefreshWithState(t *testing.T) {
	ass := assert.New(t)
	failed1 := refresh.StartRefreshScript(context.Background(), db.Get())
	failed2 := refresh.StartRefreshScript(context.Background(), db.Get())

	expected := refresh.StartRefreshScript(context.Background(), db.Get())

	failedRefresh := refresh.FetchRefreshsWithState(refresh.Prepare)
	ass.Contains(failedRefresh, *failed1)
	ass.Contains(failedRefresh, *failed2)

	time.Sleep(100 * time.Millisecond)
	runningRefresh := refresh.FetchRefreshsWithState(refresh.Running)
	getUUID := func(r refresh.Refresh) uuid.UUID { return r.UUID }
	ass.Contains(utils.Convert(runningRefresh, getUUID), expected.UUID)
}

func TestApi_OpsRefresh_StatusHandler(t *testing.T) {
	ass := assert.New(t)
	rollbackCtx, rollback := context.WithCancel(context.Background())
	current := refresh.StartRefreshScript(rollbackCtx, db.Get())
	path := "/ops/refresh/status/" + current.UUID.String()
	response := test.HTTPGet(t, path)
	rollback()

	body := test.GetBodyQuietly(response)
	ass.Equalf(http.StatusOK, response.StatusCode, "body de la réponse : %s", body)

	retour := &refresh.Refresh{}
	err := json.Unmarshal(body, &retour)
	if err != nil {
		t.Errorf("erreur survenue pendant l'unmarshalling de la réponse '%s' (cause : %s)", body, err.Error())
	}
	ass.Equal(current.UUID, retour.UUID)
}

func TestApi_OpsRefresh_ListHandler(t *testing.T) {
	ass := assert.New(t)
	test.FakeTime(t, tuTime)

	current := refresh.StartRefreshScript(context.Background(), db.Get())
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
