package stats

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"datapi/pkg/core"
)

//type AccessLogSyncer struct {
//	ctx    context.Context
//	source *pgxpool.Pool
//	target *pgxpool.Pool
//}
//
//func NewAccessLogSyncer(ctx context.Context, source *pgxpool.Pool, target *pgxpool.Pool) (*AccessLogSyncer, error) {
//	if source == nil {
//		return nil, errors.New("pas de base de données source")
//	}
//	if target == nil {
//		return nil, errors.New("pas de base de données cible")
//	}
//	err := createStructure(ctx, target)
//	return &AccessLogSyncer{
//		ctx:    ctx,
//		source: source,
//		target: target,
//	}, err
//}
//
//func NewAccessLogSyncerFromURLs(ctx context.Context, sourceURL string, targetURL string) (*AccessLogSyncer, error) {
//	if len(sourceURL) == 0 {
//		return nil, errors.New("pas d'url de base de données source")
//	}
//	sourcePool, err := pgxpool.New(ctx, sourceURL)
//	if err != nil {
//		return nil, errors.Wrap(err, "erreur pendant la lecture de l'url de la base de données source de logs")
//	}
//	if len(targetURL) == 0 {
//		return nil, errors.New("pas d'url de base de données cible")
//	}
//	targetPool, err := pgxpool.New(ctx, targetURL)
//	if err != nil {
//		return nil, errors.Wrap(err, "erreur pendant la lecture de l'url de la base de données cible de logs")
//	}
//	return NewAccessLogSyncer(ctx, sourcePool, targetPool)
//}

//// ConfigureEndpoint configure l'endpoint du package `ops`
//func (als *AccessLogSyncer) ConfigureEndpoint(endpoint *gin.RouterGroup) {
//	endpoint.GET("/sync", als.syncAccessLogsHandler)
//}
//
//func (als *AccessLogSyncer) syncAccessLogsHandler(c *gin.Context) {
//	counter, err := SyncAccessLogs(als.ctx, als.source, als.target)
//	if err != nil {
//		utils.AbortWithError(c, errors.Wrap(err, "erreur survenue pendant la synchronisation des logs"))
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{"nbLogs": strconv.Itoa(counter)})
//}

func SyncPostgresLogSaver(saverToSync *PostgresLogSaver, source *pgxpool.Pool) (int, error) {
	return syncAccessLogs(saverToSync.ctx, source, saverToSync.db)
}

func syncAccessLogs(ctx context.Context, source *pgxpool.Pool, target *pgxpool.Pool) (int, error) {
	var batch pgx.Batch

	// on crée la structure si nécessaire
	err := createStructure(ctx, target)
	if err != nil {
		return -1, errors.Wrap(err, "erreur pendant la création des tables de logs")
	}

	// on synchronize la séquence d'id des logs
	row := source.QueryRow(ctx, `select id from logs order by id desc limit 1`)
	var accessLogsCounter int
	if err := row.Scan(&accessLogsCounter); err != nil {
		return -1, errors.Wrap(err, "erreur pendant le comptage du nombre de logs depuis la source")
	}
	_, err = target.Exec(ctx, `select setval('logs_id', $1, true)`, accessLogsCounter)
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
