package stats

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

//func NewPostgresLogSaverFromConfig(ctx context.Context) (*PostgresLogSaver, error) {
//	logsDBURL := viper.GetString("logs.db_url")
//	return NewPostgresLogSaverFromURL(ctx, logsDBURL)
//}

func NewPostgresLogSaverFromURL(ctx context.Context, connexionURL string) (*PostgresLogSaver, error) {
	pool, err := pgxpool.New(ctx, connexionURL)
	if err != nil {
		return nil, errors.Wrapf(err, "erreur pendant la lecture de l'url de la base de données source '%s'", connexionURL)
	}
	return NewPostgresLogSaver(ctx, pool), nil
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
