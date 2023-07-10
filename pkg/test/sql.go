package test

import (
	"context"
	_ "embed"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

//go:embed resources/sql/insert_some_logs.sql
var insertLogsSQL string

var pool *pgxpool.Pool

func InsertSomeLogsAtTime(t time.Time) int {
	result, err := getPool().Exec(context.Background(), insertLogsSQL, t)
	if err != nil {
		log.Fatal(errors.Wrap(err, "erreur pendant l'insertion des logs"))
	}
	return int(result.RowsAffected())
}

func getPool() *pgxpool.Pool {
	if pool != nil {
		return pool
	}
	url := GetDatapiLogsDBURL()
	var err error
	pool, err = pgxpool.New(context.Background(), url)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "erreur pendant la lecture de l'url de la base de données source '%s'", url))
	}
	return pool
}
