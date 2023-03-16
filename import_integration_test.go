//go:build integration
// +build integration

package main

import (
	"github.com/signaux-faibles/datapi/src/test"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestMepHandler(t *testing.T) {
	ass := assert.New(t)
	rPath := "/utils/mep"

	response := test.HTTPGet(t, rPath)

	ass.Equalf(http.StatusOK, response.StatusCode, "body de la réponse : %s", test.GetBodyQuietly(response))
}
