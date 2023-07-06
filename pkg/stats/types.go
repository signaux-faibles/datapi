package stats

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

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

type line struct {
	date     time.Time
	path     string
	method   string
	username string
	roles    []string
}

func (l line) String() string {
	return strings.Join(l.getFieldsAsStringArray(), ";")
}

func (l line) getFieldsAsStringArray() []string {
	return []string{
		l.date.Format("20060102150405"),
		l.path,
		l.method,
		l.username,
		strings.Join(l.roles, "-"),
	}
}
