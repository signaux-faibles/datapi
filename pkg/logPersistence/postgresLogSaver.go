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

func (pgSaver *PostgresLogSaver) Initialize() error {
	return createStructure(pgSaver.ctx, pgSaver.db)
}

func (pgSaver *PostgresLogSaver) SaveLogToDB(message core.AccessLog) error {
	err := insertAccessLog(pgSaver.ctx, pgSaver.db, message)
	return err
}
