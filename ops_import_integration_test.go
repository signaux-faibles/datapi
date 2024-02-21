//go:build integration

package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"datapi/pkg/db"
	"datapi/pkg/ops/imports"
	"datapi/pkg/ops/scripts"
	"datapi/pkg/test"
)

func TestApi_OpsScripts_StartHandler(t *testing.T) {
	ass := assert.New(t)
	path := "/ops/imports/liste/refresh"
	response := test.HTTPGet(t, path)

	ass.Equalf(http.StatusOK, response.StatusCode, "body de la réponse : %s", test.GetBodyQuietly(response))
}

func TestApi_Ops_RefreshVTables(t *testing.T) {
	//utils.ConfigureLogLevel("warn")
	ass := assert.New(t)
	ctx := context.Background()

	current := scripts.StartRefreshScript(ctx, db.Get(), imports.ExecuteRefreshVTables)

	// Création d'un canal pour gérer le timeout
	timeout := time.After(10 * time.Second) // Timeout de 5 secondes
	actual := scripts.Refresh{}
	var err error
	// Boucle while avec timeout
	for actual.Status != scripts.Finished {
		time.Sleep(1 * time.Second)
		select {
		case <-timeout:
			t.Error("time out !!!")
			t.Fail()
			return
		default:
			actual, err = scripts.Fetch(current.UUID)
			require.NoError(t, err)
		}
	}

	ass.Equal(current.UUID, actual.UUID)
	ass.Equal(scripts.Finished, actual.Status)
}
