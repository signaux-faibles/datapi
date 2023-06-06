//go:build integration

package stats

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"

	"datapi/pkg/db"
)

func TestStats_syncAccessLogs(t *testing.T) {
	t.Cleanup(func() { eraseAccessLogs(sut.ctx, t, sut.db) })
	ass := assert.New(t)
	var expected, inserted, actual int
	var err error

	source := db.Get()
	target := sut.db

	// récupère le nombre de lignes de logs dans la base source
	expected = countAccessLogs(sut.ctx, t, source)

	// synchronise une première fois
	inserted, err = syncAccessLogs(sut.ctx, source, target)
	ass.NoError(err)
	ass.Equal(expected, inserted)

	// récupère le nombre
	actual = countAccessLogs(sut.ctx, t, target)
	ass.Equal(expected, actual)

	_, err = syncAccessLogs(sut.ctx, source, target)
	ass.ErrorContains(err, "ERROR: duplicate key value violates unique constraint \"logs_pkey\" (SQLSTATE 23505)")
	actual = countAccessLogs(sut.ctx, t, target)
	ass.Equal(expected, actual)
}

func countAccessLogs(ctx context.Context, t *testing.T, db *pgxpool.Pool) int {
	var counter int
	row := db.QueryRow(ctx, "SELECT COUNT(*) FROM logs")
	err := row.Scan(&counter)
	if err != nil {
		t.Errorf("erreur pendant le comptage du nombre de logs : %v", err)
	}
	return counter
}
