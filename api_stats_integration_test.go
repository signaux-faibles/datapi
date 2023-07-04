//go:build integration

package main

import (
	_ "embed"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"datapi/pkg/test"
)

func TestAPI_get_stats(t *testing.T) {
	ass := assert.New(t)
	path := "/stats/since/365/days"
	response := test.HTTPGet(t, path)

	ass.Equal(http.StatusOK, response.StatusCode)
	t.Logf("body --> %s", test.GetBodyQuietly(response))
}
