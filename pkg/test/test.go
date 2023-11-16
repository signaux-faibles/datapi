// Package test ajoute les méthodes utilisées pour les tests
package test

import (
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/jaswdr/faker"
)

var lock sync.Mutex
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

func NewFaker() faker.Faker {
	lock.Lock()
	defer lock.Unlock()
	seed := time.Now().UnixMicro()
	slog.Info("crée un nouveau faker", slog.Any("seed", seed), slog.Int("pid", os.Getpid()))
	source := rand.NewSource(seed)
	f := faker.NewWithSeed(source)
	return f
}
