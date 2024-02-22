// Package refresh : contient tout le code qui concerne l'exécution d'un `Refresh` Datapi,
// c'est-à-dire l'exécution du script sql configuré
package scripts

import (
	"context"
	_ "embed"
	"log/slog"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

var allSeparators = regexp.MustCompile(`\s+`)
var sqlComment = regexp.MustCompile(`\s*--.*\n`)

// StartRefreshScript : crée un évènement de refresh, le démarre dans une routine et retourne son UUID
func StartRefreshScript(ctx context.Context, db *pgxpool.Pool, exec Script) *Run {
	current := NewRun()
	go runScript(ctx, db, exec, current)
	return current
}

func runScript(ctx context.Context, dbase *pgxpool.Pool, exec Script, refresh *Run) {

	logger := slog.Default().With(slog.String("label", exec.Label))
	tx, err := dbase.Begin(ctx)
	if err != nil {
		refresh.fail(err.Error())
		logger.Error("Erreur à l'ouverture de la transaction", slog.Any("error", err))
		return
	}
	// Defer a rollback in case anything fails.
	defer tx.Rollback(ctx)

	for _, current := range parseSQL(exec.SQL) {
		current = allSeparators.ReplaceAllString(current, " ")
		logger.Info("démarre l'exécution", slog.Any("uuid", refresh.UUID), slog.String("sql", sqlAsLog(current)))
		refresh.run(current)
		_, err = tx.Exec(ctx, current)
		if err != nil {
			refresh.fail(err.Error())
			logger.Error(
				"Erreur à l'exécution de la requête ",
				slog.Any("error", err),
				slog.String("sql", current),
			)
			return
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		refresh.fail(err.Error())
		logger.Error("Erreur au commit de transaction", slog.Any("error", err))
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

func parseSQL(sql string) []string {
	sql = sqlComment.ReplaceAllString(sql, "")
	sql = sqlComment.ReplaceAllString(sql, "")
	return strings.Split(sql, ";\n")
}
