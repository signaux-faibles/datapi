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
	ass.Len(failedRefresh, 2)

	runningRefresh := refresh.FetchRefreshsWithState(refresh.Running)
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
	response := test.HTTPGet(t, path)

	ass.Equalf(http.StatusOK, response.StatusCode, "body de la réponse : %s", test.GetBodyQuietly(response))
}

func TestStatusHandler(t *testing.T) {
	ass := assert.New(t)
	current := refresh.StartRefreshScript(context.Background(), db.Get(), "test/refreshScript.sql")
	path := "/refresh/status/" + current.UUID.String()
	response := test.HTTPGet(t, path)

	body := test.GetBodyQuietly(response)
	ass.Equalf(http.StatusOK, response.StatusCode, "body de la réponse : %s", body)

	retour := &refresh.Refresh{}
	err := json.Unmarshal(body, &retour)
	if err != nil {
		t.Errorf("erreur survenur pendant l'unmarshalling de la réponse '%s' (cause : %s)", body, err.Error())
	}
	ass.Equal(current.UUID, retour.UUID)
}

func TestListHandler(t *testing.T) {
	ass := assert.New(t)
	current := refresh.StartRefreshScript(context.Background(), db.Get(), "erreur !!! pas de script !!!")
	expectedStatus := refresh.Failed
	rPath := "/refresh/list/" + string(expectedStatus)

	response := test.HTTPGet(t, rPath)
	body := test.GetBodyQuietly(response)
	ass.Equal(http.StatusOK, response.StatusCode, "body de la réponse : %s", body)
	actual, err := decodeRefreshArray(body)
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

func decodeRefreshArray(body []byte) ([]refresh.Refresh, error) {
	var actual []refresh.Refresh
	err := json.Unmarshal(body, &actual)
	if err != nil {
		return nil, errors.New("erreur pendant l'unmarshalling de la réponse")
	}
	return actual, nil
}
