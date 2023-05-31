package logPersistence

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"datapi/pkg/core"
)

type PostgresLogSaver struct {
	db  *pgxpool.Pool
	ctx context.Context
}

func NewPostgresLogSaver(ctx context.Context, db *pgxpool.Pool) *PostgresLogSaver {
	return &PostgresLogSaver{db: db, ctx: ctx}
}

func (pgSaver *PostgresLogSaver) SaveLogToDB(message core.LogInfos) error {
	_, err := pgSaver.db.Exec(
		pgSaver.ctx,
		`insert into logs (path, method, body, token) values ($1, $2, $3, $4);`,
		message.Path,
		message.Method,
		string(message.Body),
		message.Token[1],
	)
	return err
}
