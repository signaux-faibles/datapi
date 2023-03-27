//go:build integration
// +build integration

package main

import (
	"testing"
)

// on pourra commiter ce test quand on aura des fichiers entreprise et etablissement anonymisés
func TestMepHandler(t *testing.T) {
	//ass := assert.New(t)
	//rPath := "/utils/mep"
	//err := test.Viperize(map[string]string{
	//	"sourceEntreprise":    "/Users/raphaelsquelbut/Downloads/entreprise_sample.json.gz",
	//	"sourceEtablissement": "/Users/raphaelsquelbut/Downloads/etablissement_sample.json.gz",
	//})
	//ass.NotNil(err)
	//
	//response := test.HTTPGet(t, rPath)
	//ass.Equalf(http.StatusOK, response.StatusCode, "body de la réponse : %s", test.GetBodyQuietly(response))
}
