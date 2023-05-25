// Package refresh : contient tout le code qui concerne l'exécution d'un `Refresh` Datapi,
// c'est-à-dire l'exécution du script sql configuré
package refresh

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
	"strings"
)

// StartRefreshScript : crée un évènement de refresh, le démarre dans une routine et retourne son UUID
func StartRefreshScript(ctx context.Context, db *pgxpool.Pool, scriptPath string) *Refresh {
	current := New()
	sql, err := os.ReadFile(scriptPath)
	if err != nil {
		current.fail(err.Error())
		return current
	}
	if len(sql) <= 0 {
		current.fail("le script sql est vide")
		return current
	}
	go executeRefresh(ctx, db, string(sql), current)
	return current
}

func executeRefresh(ctx context.Context, dbase *pgxpool.Pool, sql string, refresh *Refresh) {
	tx, err := dbase.Begin(ctx)
	if err != nil {
		log.Printf("Erreur à l'ouverture de la transaction pour le refresh des vues : %s", err.Error())
		return
	}
	// Defer a rollback in case anything fails.
	defer tx.Rollback(ctx)

	for _, current := range strings.Split(sql, ";\n") {
		log.Printf("Refresh - %s - Exécute la requête : '%s'", refresh.UUID, current)
		refresh.run(current)
		_, err = tx.Exec(ctx, current)
		if err != nil {
			refresh.fail(err.Error())
			log.Printf("Erreur lors de l'exécution de la requête '%s'. Cause : %s", current, err.Error())
			return
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		refresh.fail(err.Error())
		log.Printf("Erreur lors du commit de transaction : %s", err.Error())
		return
	}
	refresh.finish()
}
