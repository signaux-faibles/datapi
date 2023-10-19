//go:build integration

package stats

import (
	"context"
	_ "embed"
	"log"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"

	"datapi/pkg/core"
	"datapi/pkg/test"
	"datapi/pkg/utils"
)

var fake = test.NewFaker()

//go:embed resources/tests/token
var realtoken string

var logSaver *PostgresLogSaver
var statsDB StatsDB
var api *API

func TestMain(m *testing.M) {
	utils.ConfigureLogLevel("debug")
	testConfig := map[string]string{}
	testConfig["postgres"] = test.GetDatapiDBURL()
	testConfig["logs.db_url"] = test.GetDatapiLogsDBURL()

	err := test.Viperize(testConfig)
	if err != nil {
		log.Panicf("Erreur pendant la Viperation de la config : %s", err)
	}

	background := context.Background()
	pool, err := pgxpool.New(background, test.GetDatapiLogsDBURL())
	if err != nil {
		log.Panicf("erreur pendant la connexion à la base de données : %s", err)
	}
	statsDB, err = createStatsDB(background, pool)
	if err != nil {
		log.Panicf("erreur pendant la création de la base de données de Stats : %s", err)
	}
	logSaver = NewPostgresLogSaver(statsDB)
	if err != nil {
		log.Panicf("erreur pendant la connexion à la base de données : %s", err)
	}
	api = NewAPI(statsDB)
	// time to API be ready
	time.Sleep(1 * time.Second)
	gin.SetMode(gin.TestMode)

	// run tests
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	// on peut placer ici du code de nettoyage si nécessaire

	os.Exit(code)
}

func TestLog_create_tables_is_idempotent(t *testing.T) {
	t.Cleanup(func() { test.EraseAccessLogs(t) })
	ass := assert.New(t)

	// on insère une ligne de logs
	// pour pouvoir vérifier qu'elle continue à exister
	expected := randomAccessLog()
	_ = logSaver.SaveLogToDB(expected)

	// on vérifie que l'initialisation est idempotente
	// `Initialize` est déjà exécuté dans `TestMain`
	err := statsDB.create()
	ass.NoError(err)

	actual, _ := getLastAccessLog(logSaver.db.ctx, logSaver.db.pool)
	ass.Equal(expected, actual)
	ass.NoError(err)
}

func TestPostgresLogSaver_SaveLogToDB(t *testing.T) {
	t.Cleanup(func() { test.EraseAccessLogs(t) })
	ass := assert.New(t)
	expected := randomAccessLog()
	err := logSaver.SaveLogToDB(expected)
	ass.NoError(err)
	actual, err := getLastAccessLog(logSaver.db.ctx, logSaver.db.pool)
	ass.NoError(err)
	ass.Equal(expected, actual)
}

//func Test_selectLines2(t *testing.T) {
//  t.Cleanup(func() { test.EraseAccessLogs(t) })
//  ass := assert.New(t)
//  var err error
//  var resultChan = make(chan row[accessLog])
//  defer close(resultChan)
//  expected := randomAccessLog()
//  err = logSaver.SaveLogToDB(expected)
//  ass.NoError(err)
//  today := time.Now()
//  tomorrow := today.AddDate(0, 0, 1)
//}

//func Test_selectLines(t *testing.T) {
//	t.Cleanup(func() { test.EraseAccessLogs(t) })
//	ass := assert.New(t)
//	var err error
//	var resultChan = make(chan row[accessLog])
//	expected := randomAccessLog()
//	err = logSaver.SaveLogToDB(expected)
//	ass.NoError(err)
//	today := time.Now()
//	tomorrow := today.AddDate(0, 0, 1)
//	go selectLogs(statsDB.ctx, statsDB.pool, today, tomorrow, resultChan)
//	var logs []accessLog
//	for l := range resultChan {
//		logs = append(logs, l.value)
//	}
//	ass.Len(logs, 1)
//	ass.Equal("christophe.ninucci@beta.gouv.fr", logs[0].username)
//	ass.Len(logs[0].roles, 109)
//	ass.Equal(expected.Method, logs[0].method)
//	ass.Equal(expected.Path, logs[0].path)
//	ass.Contains(logs[0].roles, "score")
//	ass.Contains(logs[0].roles, "dgefp")
//	ass.Contains(logs[0].roles, "bdf")
//	ass.Contains(logs[0].roles, "urssaf")
//	ass.Contains(logs[0].roles, "France entière")
//	ass.Contains(logs[0].roles, "01")
//	ass.Contains(logs[0].roles, "16")
//	ass.Contains(logs[0].roles, "2A")
//	ass.Contains(logs[0].roles, "972")
//}

func randomAccessLog() core.AccessLog {
	random := core.AccessLog{
		Path:   fake.File().AbsoluteFilePathForUnix(2),
		Method: fake.Internet().HTTPMethod(),
		Body:   fake.Lorem().Bytes(256),
		Token:  realtoken,
	}
	return random
}
