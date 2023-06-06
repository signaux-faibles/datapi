package stats

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"datapi/pkg/core"
)

func syncAccessLogs(ctx context.Context, source *pgxpool.Pool, target *pgxpool.Pool) (int, error) {
	var batch pgx.Batch

	// on synchronize la séquence d'id des logs
	row := source.QueryRow(ctx, `select id from logs order by id desc limit 1`)
	var accessLogsCounter int
	if err := row.Scan(&accessLogsCounter); err != nil {
		return -1, errors.Wrap(err, "erreur pendant le comptage du nombre de logs depuis la source")
	}
	_, err := target.Exec(ctx, `select setval('logs_id', $1, true)`, accessLogsCounter)
	if err != nil {
		return -1, errors.Wrap(err, "erreur pendant la mise à jour des log_id dans la cible")
	}

	// on synchronise les logs
	rows, err := source.Query(ctx, `select id, date_add, path, method, body, token from logs`)
	if err != nil {
		return -1, errors.Wrap(err, "erreur pendant la récupération des logs depuis la source")
	}
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
		return -1, errors.Wrap(err, "erreur pendant l'insertion des access logs vers la cible")
	}
	return batch.Len(), nil
}
