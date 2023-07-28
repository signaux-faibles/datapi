package stats

import (
	"context"

	"github.com/pkg/errors"

	"datapi/pkg/core"
)

type PostgresLogSaver struct {
	db StatsDB
}

func NewPostgresLogSaver(db StatsDB) *PostgresLogSaver {
	return &PostgresLogSaver{db: db}
}

func NewPostgresLogSaverFromURL(ctx context.Context, connexionURL string) (*PostgresLogSaver, error) {
	statsDB, err := createStatsDBFromURL(ctx, connexionURL)
	if err != nil {
		return nil, errors.Wrap(err, "erreur lors de l'initialisation du log saver")
	}
	return NewPostgresLogSaver(statsDB), nil
}

func (pgSaver *PostgresLogSaver) SaveLogToDB(message core.AccessLog) error {
	err := insertAccessLog(pgSaver.db.ctx, pgSaver.db.pool, message)
	if err != nil {
		return errors.Wrap(err, "erreur pendant la sauvegarde de l'access log")
	}
	return nil
}
