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

func executeRefresh(ctx context.Context, db *pgxpool.Pool, sql string, refresh *Refresh) {
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatalf("Erreur à l'ouverture de la transaction pour le refresh des vues : %s", err.Error())
	}
	for _, current := range strings.Split(sql, ";\n") {
		log.Printf("Refresh - %s - Exécute la requête : '%s'", refresh.UUID, current)
		refresh.run(current)
		_, err = tx.Exec(ctx, current)
		if err != nil {
			refresh.fail(err.Error())
			err := tx.Rollback(ctx)
			if err != nil {
				log.Fatalf("Erreur lors du rollback de transaction. Cause : %s", err.Error())
			}
			log.Fatalf("Erreur lors de l'exécution de la requête '%s'. Cause : %s", current, err.Error())
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		refresh.fail(err.Error())
		log.Fatalf("Erreur lors du commit de transaction : %s", err.Error())
	}
	refresh.finish()
}
