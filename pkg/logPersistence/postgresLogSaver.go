package logPersistence

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

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
	err := createStructure(pgSaver.ctx, pgSaver.db)
	if err != nil {
		return errors.Wrap(err, "erreur pendant l'initialisation de la base de logs")
	}
	return nil
}

func (pgSaver *PostgresLogSaver) SaveLogToDB(message core.AccessLog) error {
	err := insertAccessLog(pgSaver.ctx, pgSaver.db, message)
	if err != nil {
		return errors.Wrap(err, "erreur pendant la sauvegarde de l'access log")
	}
	return nil
}
