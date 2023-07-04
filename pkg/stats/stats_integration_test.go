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
var realtoken = "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJtb3hEcWt2QzRVMy1LVUVLNmJoanJ0SzdPRWNlbzlqejhnNERnS0JQeWo0In0.eyJleHAiOjE2ODQ1NzYxOTQsImlhdCI6MTY4NDU3NTI5NCwiYXV0aF90aW1lIjoxNjg0NTc1MjkzLCJqdGkiOiIzNjJmYmM0Ni0wYjczLTRmZWItOWM0ZS0wYzczNzA3MjJlMmQiLCJpc3MiOiJodHRwczovL3Rlc3Rpbmcuc2lnbmF1eC1mYWlibGVzLmRldi9hdXRoL3JlYWxtcy9tYXN0ZXIiLCJhdWQiOiJhY2NvdW50Iiwic3ViIjoiODljYWIzMDgtMmM1YS00MDRmLTlhZDktM2E1MWM3MWNjZGE4IiwidHlwIjoiQmVhcmVyIiwiYXpwIjoic2lnbmF1eGZhaWJsZXMiLCJub25jZSI6ImNhN2I4MTY4LTM2NWEtNDZkMC04ZGFjLTZjMzE5OGFiNWM0MyIsInNlc3Npb25fc3RhdGUiOiIyYjEyYTgzMi1kOWE5LTQ3MjYtYTc0Yi04NDk4Nzc1OWZmZmEiLCJhY3IiOiIxIiwiYWxsb3dlZC1vcmlnaW5zIjpbImh0dHA6Ly9sb2NhbGhvc3Q6ODA4MCIsImh0dHBzOi8vdGVzdGluZy5zaWduYXV4LWZhaWJsZXMuZGV2Il0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJkZWZhdWx0LXJvbGVzLW1hc3RlciIsIm9mZmxpbmVfYWNjZXNzIiwidW1hX2F1dGhvcml6YXRpb24iXX0sInJlc291cmNlX2FjY2VzcyI6eyJzaWduYXV4ZmFpYmxlcyI6eyJyb2xlcyI6WyI4OCIsIjg5IiwiMDEiLCIwMiIsIjAzIiwiMDQiLCIwNSIsIjA2IiwiMDciLCIwOCIsIjA5Iiwic2NvcmUiLCI5MCIsIjkxIiwiOTIiLCI5MyIsIjk0IiwiOTUiLCIxMCIsIjExIiwiMTIiLCIxMyIsIjE0IiwiMTUiLCIxNiIsIjE3IiwiMTgiLCIxOSIsImRnZWZwIiwiMjEiLCIyMiIsIjIzIiwiMjQiLCIyNSIsIjI2IiwiMjciLCIyOCIsIjI5IiwicGdlIiwiMkEiLCIyQiIsIjMwIiwiMzEiLCIzMiIsIjMzIiwiMzQiLCJkZXRlY3Rpb24iLCIzNSIsIjM2IiwiMzciLCJiZGYiLCIzOCIsIjM5IiwiNDAiLCI0MSIsIjQyIiwiNDMiLCI0NCIsInVyc3NhZiIsIjQ1IiwiNDYiLCI0NyIsIjQ4IiwiNDkiLCI1MCIsIjUxIiwiNTIiLCI1MyIsIjU0IiwiNTUiLCI1NiIsIjU3IiwiNTgiLCI1OSIsIkZyYW5jZSBlbnRpw6hyZSIsIjYwIiwiNjEiLCI2MiIsIjYzIiwiNjQiLCI2NSIsIjY2IiwiNjciLCI2OCIsIjY5IiwiOTcxIiwiOTcyIiwiOTczIiwiOTc0IiwiOTc2IiwiNzAiLCI3MSIsIjcyIiwiNzMiLCI3NCIsIjc1IiwiNzYiLCI3NyIsIjc4IiwiNzkiLCJ3ZWthbiIsIjgwIiwiODEiLCI4MiIsIjgzIiwiODQiLCI4NSIsIjg2IiwiODciXX0sImFjY291bnQiOnsicm9sZXMiOlsibWFuYWdlLWFjY291bnQiLCJtYW5hZ2UtYWNjb3VudC1saW5rcyIsInZpZXctcHJvZmlsZSJdfX0sInNjb3BlIjoib3BlbmlkIGVtYWlsIHByb2ZpbGUiLCJzaWQiOiIyYjEyYTgzMi1kOWE5LTQ3MjYtYTc0Yi04NDk4Nzc1OWZmZmEiLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwibmFtZSI6IkNocmlzdG9waGUgTklOVUNDSSIsInByZWZlcnJlZF91c2VybmFtZSI6ImNocmlzdG9waGUubmludWNjaUBiZXRhLmdvdXYuZnIiLCJnaXZlbl9uYW1lIjoiQ2hyaXN0b3BoZSIsImZhbWlseV9uYW1lIjoiTklOVUNDSSIsImVtYWlsIjoiY2hyaXN0b3BoZS5uaW51Y2NpQGJldGEuZ291di5mciJ9.Ynbet0C3jUOv-Sy4ul0Rbgrt-qCx3GS0E89E-vA-NqMovficTwavDzybxh_RfjsG2omNSeuueLJwVU5nplwO-ohm09o7TveP1pehy1GeFG1LYXP2MK68OaD9WQo5DfUtrFUtRsvy_K9aEV_cM24MjN689tGeo-FerLobZNUj0cmnsLrfuZwLe2V7TxZIzaEkX2i96jzlEbtuGfCOimTTiw5lpEkiJPBaJFm1BsvmMdogMKQeBgJA9Fm04__gSf53RopVqIP2ZtfjuiuvxA2-83ZM0h1iej150v6RH8PwZX3RGsvBtCzXrYPCy_4L0oU5Jl2-tyApoUveI5bMuKIr2w"

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

func Test_selectLines(t *testing.T) {
	t.Cleanup(func() { eraseAccessLogs(sut.ctx, t, sut.db) })
	ass := assert.New(t)
	expected := randomAccessLog()
	err := sut.SaveLogToDB(expected)
	ass.NoError(err)
	oneYearAgo := tuTime.Add(time.Duration(-365*24) * time.Hour)
	logs, err := selectLogs(sut.ctx, sut.db, oneYearAgo)
	ass.NoError(err)
	ass.Len(logs, 1)
	ass.Equal("christophe.ninucci@beta.gouv.fr", logs[0].username)
	ass.Len(logs[0].roles, 109)
	ass.Equal(expected.Method, logs[0].method)
	ass.Equal(expected.Path, logs[0].path)
	ass.Contains(logs[0].roles, "score")
	ass.Contains(logs[0].roles, "dgefp")
	ass.Contains(logs[0].roles, "bdf")
	ass.Contains(logs[0].roles, "urssaf")
	ass.Contains(logs[0].roles, "France entière")
	ass.Contains(logs[0].roles, "01")
	ass.Contains(logs[0].roles, "16")
	ass.Contains(logs[0].roles, "2A")
	ass.Contains(logs[0].roles, "972")
}

func randomAccessLog() core.AccessLog {
	random := core.AccessLog{
		Path:   fake.File().AbsoluteFilePathForUnix(2),
		Method: fake.Internet().HTTPMethod(),
		Body:   fake.Lorem().Bytes(256),
		Token:  realtoken,
	}
	return random
}

func eraseAccessLogs(ctx context.Context, t *testing.T, db *pgxpool.Pool) {
	_, err := db.Exec(ctx, "DELETE FROM logs")
	if err != nil {
		t.Errorf("erreur pendant la suppression des access logs : %v", err)
	}
}
