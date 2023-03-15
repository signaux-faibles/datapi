//go:build integration
// +build integration

package main

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/signaux-faibles/datapi/src/refresh"
	"github.com/signaux-faibles/datapi/src/test"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestFetchNonExistentRefresh(t *testing.T) {
	ass := assert.New(t)
	result, err := refresh.Fetch(uuid.New())
	ass.Equal(refresh.Empty, result)
	ass.NotNil(err)
}

func TestRunRefreshScript(t *testing.T) {
	ass := assert.New(t)
	current := refresh.StartRefreshScript(context.Background(), db.Get(), "test/refreshScript.sql")
	t.Logf("refreshID is running with id : %s", current.UUID)
	result, err := refresh.Fetch(current.UUID)
	ass.Nil(err)
	ass.NotNil(result)
}

func TestLastRefreshState(t *testing.T) {
	ass := assert.New(t)
	lastRefreshState := refresh.FetchLast()
	if lastRefreshState == refresh.Empty {
		refresh.StartRefreshScript(context.Background(), db.Get(), "test/refreshScript.sql")
	}
	time.Sleep(100 * time.Millisecond)
	lastRefreshState = refresh.FetchLast()
	ass.NotEmpty(lastRefreshState)
	t.Logf("Description du dernier refresh : %s", lastRefreshState)
}

func TestFetchRefreshWithState(t *testing.T) {
	ass := assert.New(t)
	refresh.StartRefreshScript(context.Background(), db.Get(), "test/script qui n'existe pas")
	refresh.StartRefreshScript(context.Background(), db.Get(), "test/autre script foireux")
	expected := refresh.StartRefreshScript(context.Background(), db.Get(), "test/refreshScript.sql")
	time.Sleep(100 * time.Millisecond)

	failedRefresh := refresh.FetchRefreshsWithState(refresh.Failed)
	t.Logf("Description des refresh erronés : %s", failedRefresh)
	ass.Len(failedRefresh, 2)

	runningRefresh := refresh.FetchRefreshsWithState(refresh.Running)
	t.Logf("Description du refresh en cours : %s", runningRefresh)
	ass.Condition(func() bool {
		for _, current := range runningRefresh {
			if current.UUID == expected.UUID {
				return true
			}
		}
		return false
	})
}

func TestStartHandler(t *testing.T) {
	ass := assert.New(t)
	path := "/refresh/start"
	response, body, _ := test.HTTPGetAndFormatBody(t, path)

	t.Logf("response: %s", body)
	ass.Equal(http.StatusOK, response.StatusCode)
}

func TestStatusHandler(t *testing.T) {
	ass := assert.New(t)
	current := refresh.StartRefreshScript(context.Background(), db.Get(), "test/refreshScript.sql")
	path := "/refresh/status/" + current.UUID.String()
	response := test.HTTPGet(t, path)

	ass.Equal(http.StatusOK, response.StatusCode)
	retour := &refresh.Refresh{}
	body := json.NewDecoder(response.Body)
	body.Decode(retour)
	t.Logf("expected : %s", current)
	t.Logf("actual : %s", retour)
	ass.Equal(current.UUID, retour.UUID)
}

func TestListHandler(t *testing.T) {
	ass := assert.New(t)
	current := refresh.StartRefreshScript(context.Background(), db.Get(), "erreur !!! pas de script !!!")
	expectedStatus := refresh.Failed
	rPath := "/refresh/list/" + string(expectedStatus)

	response := test.HTTPGet(t, rPath)

	ass.Equal(http.StatusOK, response.StatusCode)
	actual, err := decodeRefreshArray(response)
	ass.Nil(err)
	ass.GreaterOrEqual(len(actual), 1)
	ass.Conditionf(func() bool {
		// tous les refresh doivent avoir le status `failed`
		for _, current := range actual {
			if current.Status != expectedStatus {
				return false
			}

		}
		return true
	}, "Un des refresh n'a pas le status `failed`")
	ass.Conditionf(func() bool {
		for _, current := range actual {
			if current.UUID == current.UUID {
				return true
			}
		}
		return false
	}, "Le refresh avec l'uuid %s n'existe pas", current)
}

func decodeRefreshArray(response *http.Response) ([]refresh.Refresh, error) {
	var actual []refresh.Refresh
	all, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New("erreur pendant la lecture de la réponse")
	}
	err = json.Unmarshal(all, &actual)
	if err != nil {
		return nil, errors.New("erreur pendant l'unmarshalling de la réponse")
	}
	return actual, nil
}
