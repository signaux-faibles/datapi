// Package test ajoute les méthodes utilisées pour les tests
package test

import (
	"os"
	"path"
	"runtime"
)

// cette fonction se lance dès que le package `test` est chargé
// ainsi, dès qu'un test se lance, le répertoire de travail est `datapi`
func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}
