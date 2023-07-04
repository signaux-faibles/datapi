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
	return err
}

func insertAccessLog(ctx context.Context, db *pgxpool.Pool, message core.AccessLog) error {
	_, err := db.Exec(
		ctx,
		`insert into logs (path, method, body, token) values ($1, $2, $3, $4) ;`,
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

func selectLogs(ctx context.Context, dbPool *pgxpool.Pool, since time.Time, to time.Time) ([]line, error) {
	stats := []line{}
	rows, err := dbPool.Query(ctx, selectLogsSQL, since.Truncate(day), to.Truncate(day))
	if err != nil {
		return stats, errors.Wrap(err, "erreur pendant la requête de sélection des logs")
	}
	for rows.Next() {
		var statLine line
		err := rows.Scan(&statLine.date, &statLine.path, &statLine.method, &statLine.username, &statLine.roles)
		if err != nil {
			return stats, errors.Wrap(err, "erreur pendant la récupération des résultats")
		}
		stats = append(stats, statLine)
	}
	return stats, nil
}
