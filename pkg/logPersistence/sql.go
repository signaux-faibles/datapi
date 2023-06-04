package logPersistence

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"datapi/pkg/core"
)

//go:embed sql/create_or_replace.sql
var sql string

func createStructure(ctx context.Context, db *pgxpool.Pool) error {
	err := db.BeginFunc(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, sql)
		return err
	})
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
