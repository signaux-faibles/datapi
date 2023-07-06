package test

import (
	"context"
	_ "embed"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

//go:embed resources/sql/insert_some_logs.sql
var insertLogsSQL string

func InsertSomeLogsAtTime(t time.Time) int {
	url := GetDatapiLogsDBURL()
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "erreur pendant la lecture de l'url de la base de données source '%s'", url))
	}
	var counter int
	var insertLogSQL string
	for counter, insertLogSQL = range strings.Split(insertLogsSQL, "\n") {
		if len(insertLogSQL) == 0 || strings.HasPrefix(insertLogSQL, "--") {
			continue
		}
		_, err = pool.Exec(context.Background(), insertLogSQL, t)
		if err != nil {
			log.Fatal(errors.Wrap(err, "erreur pendant l'insertion des logs"))
		}
	}
	return counter
}
