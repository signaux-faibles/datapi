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
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"datapi/pkg/core"
	"datapi/pkg/db"
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

	test.Wait4PostgresIsReady(test.GetDatapiDBURL())
	test.Wait4PostgresIsReady(test.GetDatapiLogDBURL())

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
	t.Cleanup(func() {
		err := eraseAccessLogs(sut.ctx, sut.db)
		if err != nil {
			t.Error("erreur pendant le nettoyage :", err)
		}
	})
	ass := assert.New(t)
	expected := randomAccessLog()
	err := sut.SaveLogToDB(expected)
	ass.NoError(err)
	actual, err := getLastAccessLog(sut.ctx, sut.db)
	ass.NoError(err)
	ass.Equal(expected, actual)
}

func TestPostgresLogSaver_syncAccessLogs(t *testing.T) {
	t.Cleanup(func() {
		err := eraseAccessLogs(sut.ctx, sut.db)
		if err != nil {
			t.Error("erreur pendant le nettoyage :", err)
		}
	})
	ass := assert.New(t)
	var expected, inserted, actual int
	var err error

	source := db.Get()
	target := sut.db

	// récupère le nombre de lignes de logs dans la base source
	expected, err = countAccessLogs(sut.ctx, source)
	ass.NoError(err)

	// synchronise une première fois
	inserted, err = syncAccessLogs(sut.ctx, source, target)
	ass.NoError(err)
	ass.Equal(expected, inserted)

	// récupère le nombre
	actual, err = countAccessLogs(sut.ctx, target)
	ass.NoError(err)
	ass.Equal(expected, actual)

	_, err = syncAccessLogs(sut.ctx, source, target)
	ass.ErrorContains(err, "ERROR: duplicate key value violates unique constraint \"logs_pkey\" (SQLSTATE 23505)")
	actual, err = countAccessLogs(sut.ctx, target)
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

func eraseAccessLogs(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, "DELETE FROM logs")
	if err != nil {
		return errors.Wrap(err, "erreur pendant la suppression des access logs")
	}
	return nil
}
