package logPersistence

import (
	"context"
	_ "embed"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"datapi/pkg/core"
)

//go:embed sql/create_or_replace.sql
var sql string

func createStructure(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, sql)
	return err
}

func insertAccessLog(ctx context.Context, db *pgxpool.Pool, message core.AccessLog) error {
	_, err := db.Exec(
		ctx,
		`insert into logs (path, method, body, token) values ($1, $2, $3, $4);`,
		message.Path,
		message.Method,
		string(message.Body),
		message.Token,
	)
	return err
}

func getLastAccessLog(ctx context.Context, db *pgxpool.Pool) (core.AccessLog, error) {
	var r core.AccessLog
	row := db.QueryRow(ctx, "SELECT path, method, body, token FROM logs ORDER BY date_add DESC LIMIT 1;")
	err := row.Scan(&r.Path, &r.Method, &r.Body, &r.Token)
	return r, err
}

func syncAccessLogs(ctx context.Context, source *pgxpool.Pool, target *pgxpool.Pool) (int, error) {
	rows, err := source.Query(ctx, `select id, date_add, path, method, body, token from logs`)
	if err != nil {
		return -1, errors.Wrap(err, "erreur pendant la récupération des logs depuis la source")
	}
	var batch pgx.Batch
	for rows.Next() {
		var currentID int
		var currentTime time.Time
		var currentLog core.AccessLog
		err := rows.Scan(&currentID, &currentTime, &currentLog.Path, &currentLog.Method, &currentLog.Body, &currentLog.Token)
		if err != nil {
			return -1, errors.Wrap(err, "erreur pendant la lecture d'une log")
		}
		batch.Queue(`insert into logs (id, date_add, path, method, body, token) values ($1, $2 ,$3, $4, $5, $6)`,
			currentID, currentTime, currentLog.Path, currentLog.Method, string(currentLog.Body), currentLog.Token)
	}
	br := target.SendBatch(ctx, &batch)
	_, err = br.Exec()
	if err != nil {
		return -1, errors.Wrap(err, "erreur pendant l'insertion logs vers la cible")
	}
	return batch.Len(), nil
}

func countAccessLogs(ctx context.Context, db *pgxpool.Pool) (int, error) {
	var counter int
	row := db.QueryRow(ctx, "SELECT COUNT(*) FROM logs")
	err := row.Scan(&counter)
	if err != nil {
		return -1, errors.Wrap(err, "erreur pendant le comptage des access logs")
	}
	return counter, nil
}

func eraseAccessLogs(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, "DELETE FROM logs")
	if err != nil {
		return errors.Wrap(err, "erreur pendant la suppression des access logs")
	}
	return nil
}
