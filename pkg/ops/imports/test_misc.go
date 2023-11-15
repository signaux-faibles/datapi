package imports

import (
	"path/filepath"
	"runtime"
	"testing"
)

func pointer[T any](value T) *T {
	return &value
}

func buildPath(t *testing.T, path string) string {
	_, testFilename, _, _ := runtime.Caller(1)
	base := filepath.Dir(testFilename)
	abs := filepath.Join(base, path)
	t.Log("chemin absolu de la resource de test", abs)
	return abs
}
