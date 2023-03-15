//go:build integration
// +build integration

package main

import (
	"github.com/signaux-faibles/datapi/src/test"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestMepHandler(t *testing.T) {
	ass := assert.New(t)
	rPath := "/utils/mep"

	response := test.HTTPGet(t, rPath)
	body := new(strings.Builder)
	if _, err := io.Copy(body, response.Body); err != nil {
		t.Errorf("erreur pendant la copie de la response : %s", err)
		t.Fail()
	}
	ass.Equalf(http.StatusOK, response.StatusCode, "body de la réponse : %s", body)

}
