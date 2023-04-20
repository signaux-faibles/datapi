//go:build integration

package main

import (
	"github.com/signaux-faibles/datapi/src/test"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestApi_Wekan_endpoint_is_configured(t *testing.T) {
	ass := assert.New(t)
	path := "/kanban/config"
	response := test.HTTPGet(t, path)

	ass.Equal(http.StatusOK, response.StatusCode)
}
