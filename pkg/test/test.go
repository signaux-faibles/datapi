// Package test ajoute les méthodes utilisées pour les tests
package test

import (
	"log/slog"
	"path/filepath"
	"runtime"
)

var projectPath string

func init() {
	_, here, _, _ := runtime.Caller(0)
	projectPath = filepath.Join(here, "..", "..", "..")
	slog.Info("le chemin du projet est initialisé", slog.String("path", projectPath))
}

func ProjectPathOf(relativePath string) string {
	ppath := filepath.Join(projectPath, relativePath)
	slog.Debug("chemin de la resource de test", slog.String("path", ppath))
	return ppath
}
