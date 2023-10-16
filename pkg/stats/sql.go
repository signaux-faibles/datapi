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

//go:embed resources/sql/select_activite_par_utilisateur.sql
var selectActiviteParUtilisateurSQL string

//go:embed resources/sql/select_activite_par_jour.sql
var selectActiviteParJourSQL string

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

func selectLogs(ctx context.Context, dbPool *pgxpool.Pool, since time.Time, to time.Time, r chan accessLog) {
	defer close(r)
	rows, err := dbPool.Query(ctx, selectLogsSQL, since.Truncate(day), to.Truncate(day))
	if err != nil {
		r <- accessLog{err: errors.Wrap(err, "erreur pendant la requête de sélection des logs")}
	}
	for rows.Next() {
		var statLine accessLog
		err := rows.Scan(&statLine.date, &statLine.path, &statLine.method, &statLine.username, &statLine.segment, &statLine.roles)
		if err != nil {
			r <- accessLog{err: errors.Wrap(err, "erreur pendant la récupération des résultats")}
		}
		r <- statLine
	}
	if err := rows.Err(); err != nil {
		r <- accessLog{err: errors.Wrap(err, "erreur après la récupération des résultats")}
	}
}

func selectActiviteParUtilisateur(
	ctx context.Context,
	dbPool *pgxpool.Pool,
	since time.Time,
	to time.Time,
	r chan activiteParUtilisateur,
) {
	defer close(r)
	rows, err := dbPool.Query(ctx, selectActiviteParUtilisateurSQL, since.Truncate(day), to.Truncate(day))
	if err != nil {
		r <- activiteParUtilisateur{err: errors.Wrap(err, "erreur pendant la requête de sélection des activite")}
	}
	for rows.Next() {
		var activite activiteParUtilisateur
		err := rows.Scan(&activite.username, &activite.visites, &activite.actions, &activite.segment)
		if err != nil {
			r <- activiteParUtilisateur{err: errors.Wrap(err, "erreur pendant la récupération des résultats")}
		}
		r <- activite
	}
	if err := rows.Err(); err != nil {
		r <- activiteParUtilisateur{err: errors.Wrap(err, "erreur après la récupération des résultats")}
	}
}

func selectActiviteParJour(
	ctx context.Context,
	dbPool *pgxpool.Pool,
	since time.Time,
	to time.Time,
	r chan activiteParJour,
) {
	defer close(r)
	rows, err := dbPool.Query(ctx, selectActiviteParJourSQL, since.Truncate(day), to.Truncate(day))
	if err != nil {
		r <- activiteParJour{err: errors.Wrap(err, "erreur pendant la requête de sélection des activite/jour")}
	}
	for rows.Next() {
		var activite activiteParJour
		err := rows.Scan(
			&activite.jour,
			&activite.username,
			&activite.actions,
			&activite.recherches,
			&activite.fiches,
			&activite.segment,
		)
		if err != nil {
			r <- activiteParJour{err: errors.Wrap(err, "erreur pendant la récupération des résultats/jour")}
		}
		r <- activite
	}
	if err := rows.Err(); err != nil {
		r <- activiteParJour{err: errors.Wrap(err, "erreur après la récupération des résultats/jour")}
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
