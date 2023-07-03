//go:build integration

package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"datapi/pkg/test"
)

func TestAPI_get_stats(t *testing.T) {
	ass := assert.New(t)
	path := "/stats/fetchAll"
	response := test.HTTPGet(t, path)

	ass.Equalf(http.StatusOK, response.StatusCode, "body de la réponse : %s", test.GetBodyQuietly(response))
}
