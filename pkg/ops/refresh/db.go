// Package refresh : contient tout le code qui concerne l'exécution d'un `Refresh` Datapi,
// c'est-à-dire l'exécution du script sql configuré
package refresh

import (
	"context"
	_ "embed"
	"log/slog"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/populate_v_tables.sql
var sqlPopulateVTables string

// StartRefreshScript : crée un évènement de refresh, le démarre dans une routine et retourne son UUID
func StartRefreshScript(ctx context.Context, db *pgxpool.Pool) *Refresh {
	current := New()
	go executeRefresh(ctx, db, sqlPopulateVTables, current)
	return current
}

func executeRefresh(ctx context.Context, dbase *pgxpool.Pool, sql string, refresh *Refresh) {
	allSeparators := regexp.MustCompile(`\s+`)
	tx, err := dbase.Begin(ctx)
	if err != nil {
		slog.Error("Erreur à l'ouverture de la transaction pour mise à jour des v tables", slog.Any("error", err))
		return
	}
	// Defer a rollback in case anything fails.
	defer tx.Rollback(ctx)

	for _, current := range strings.Split(sql, ";\n") {
		current = allSeparators.ReplaceAllString(current, " ")
		slog.Info("mise à jour des v tables", slog.Any("uuid", refresh.UUID), slog.String("sql", sqlAsLog(current)))
		refresh.run(current)
		_, err = tx.Exec(ctx, current)
		if err != nil {
			refresh.fail(err.Error())
			slog.Error(
				"Erreur à l'ouverture de l'exécution de la requête de mise à jour des v tables",
				slog.Any("error", err),
				slog.String("sql", current),
			)
			return
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		refresh.fail(err.Error())
		slog.Error("Erreur à l'ouverture du commit de transaction de mise à jour des v tables", slog.Any("error", err))
		return
	}
	refresh.finish()
}

func sqlAsLog(sql string) string {
	if len(sql) < 50 {
		return sql
	}
	return sql[:47] + "..."
}
