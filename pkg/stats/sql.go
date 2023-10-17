package stats

import (
	"context"
	_ "embed"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"datapi/pkg/core"
)

//go:embed resources/sql/create_tables_and_views.sql
var createTablesSQL string

//go:embed resources/sql/select_logs.sql
var selectLogsSQL string

const day = time.Duration(24) * time.Hour

func createStructure(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, createTablesSQL)
	return errors.Wrap(err, "erreur lors de l'initialisation de la base de logs/stats")
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
	return errors.Wrap(err, "erreur lors de l'insertion d'un access log ")
}

func getLastAccessLog(ctx context.Context, db *pgxpool.Pool) (core.AccessLog, error) {
	var r core.AccessLog
	row := db.QueryRow(ctx, "SELECT path, method, body, token FROM logs ORDER BY date_add DESC LIMIT 1;")
	err := row.Scan(&r.Path, &r.Method, &r.Body, &r.Token)
	return r, err
}

func selectLogs(ctx context.Context, dbPool *pgxpool.Pool, since time.Time, to time.Time, r chan row[accessLog]) {
	defer close(r)
	rows, err := dbPool.Query(ctx, selectLogsSQL, since.Truncate(day), to.Truncate(day))
	if err != nil {
		r <- rowWithError(accessLog{}, errors.Wrap(err, "erreur pendant la requête de sélection des logs"))
	}
	for rows.Next() {
		var statLine accessLog
		err := rows.Scan(&statLine.date, &statLine.path, &statLine.method, &statLine.username, &statLine.segment, &statLine.roles)
		if err != nil {
			r <- rowWithError(accessLog{}, errors.Wrap(err, "erreur pendant la récupération des résultats"))
		}
		r <- newRow(statLine)
	}
	if err := rows.Err(); err != nil {
		r <- rowWithError(accessLog{}, errors.Wrap(err, "erreur après la récupération des résultats"))
	}
}

type StatsDB struct {
	pool *pgxpool.Pool
	ctx  context.Context
}

func createStatsDB(ctx context.Context, db *pgxpool.Pool) (StatsDB, error) {
	err := createStructure(ctx, db)
	if err != nil {
		return StatsDB{}, errors.Wrap(err, "erreur lors de la création de la base de données de stats")
	}
	return StatsDB{pool: db, ctx: ctx}, nil
}

func createStatsDBFromURL(ctx context.Context, connexionURL string) (StatsDB, error) {
	pool, err := pgxpool.New(ctx, connexionURL)
	if err != nil {
		return StatsDB{}, errors.Wrapf(err, "erreur pendant la lecture de l'url de la base de données source '%s'", connexionURL)
	}
	return createStatsDB(ctx, pool)
}

func (db StatsDB) create() error {
	return createStructure(db.ctx, db.pool)
}
