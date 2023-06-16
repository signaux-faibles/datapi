//go:build integration

package stats

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"

	"datapi/pkg/core"
	"datapi/pkg/test"
)

var tuTime = time.Date(2023, 03, 10, 17, 41, 58, 651387237, time.UTC)
var fake faker.Faker

// SYSTEM UNDER TEST
var sut *PostgresLogSaver

func init() {
	fake = faker.New()
}

func TestMain(m *testing.M) {
	var err error
	testConfig := map[string]string{}
	testConfig["postgres"] = test.GetDatapiDBURL()
	testConfig["logs.db_url"] = test.GetDatapiLogDBURL()

	err = test.Viperize(testConfig)
	if err != nil {
		log.Printf("Erreur pendant la Viperation de la config : %s", err)
	}

	background := context.Background()
	pool, err := pgxpool.New(background, test.GetDatapiLogDBURL())
	if err != nil {
		log.Fatalf("erreur pendant la connexion à la base de données : %s", err)
	}
	sut = NewPostgresLogSaver(background, pool)
	err = sut.Initialize()
	if err != nil {
		log.Fatalf("erreur pendant la connexion à la base de données : %s", err)
	}

	// time to API be ready
	time.Sleep(1 * time.Second)
	// run tests
	test.FakeTime(tuTime)
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	// on peut placer ici du code de nettoyage si nécessaire

	test.UnfakeTime()
	os.Exit(code)
}

func TestLog_Initialize_idempotent(t *testing.T) {
	t.Cleanup(func() { eraseAccessLogs(sut.ctx, t, sut.db) })
	ass := assert.New(t)

	// on insère une ligne de logs
	expected := randomAccessLog()
	_ = sut.SaveLogToDB(expected)

	// on vérifie que l'initialisation est idempotente
	// `Initialize` est déjà exécuté dans `TestMain`
	err := sut.Initialize()
	ass.NoError(err)

	actual, _ := getLastAccessLog(sut.ctx, sut.db)
	ass.Equal(expected, actual)
	ass.NoError(err)
}

func TestPostgresLogSaver_SaveLogToDB(t *testing.T) {
	t.Cleanup(func() { eraseAccessLogs(sut.ctx, t, sut.db) })
	ass := assert.New(t)
	expected := randomAccessLog()
	err := sut.SaveLogToDB(expected)
	ass.NoError(err)
	actual, err := getLastAccessLog(sut.ctx, sut.db)
	ass.NoError(err)
	ass.Equal(expected, actual)
}

func randomAccessLog() core.AccessLog {
	random := core.AccessLog{
		Path:   fake.File().AbsoluteFilePathForUnix(2),
		Method: fake.Internet().HTTPMethod(),
		Body:   fake.Lorem().Bytes(256),
		Token:  fake.Internet().User(),
	}
	return random
}

func eraseAccessLogs(ctx context.Context, t *testing.T, db *pgxpool.Pool) {
	_, err := db.Exec(ctx, "DELETE FROM logs")
	if err != nil {
		t.Errorf("erreur pendant la suppression des access logs : %v", err)
	}
}
