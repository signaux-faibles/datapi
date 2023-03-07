package refresh

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
	"strings"
)

func ExecRefreshScript(ctx context.Context, db *pgxpool.Pool, scriptPath string) uuid.UUID {
	current := New(uuid.New())
	sql, err := os.ReadFile(scriptPath)
	if err != nil {
		current.Fail(err.Error())
		return current.Id
	}
	if len(sql) <= 0 {
		current.Fail("le script sql est vide")
		return current.Id
	}
	go executeRefresh(ctx, db, string(sql), current)
	return current.Id
}

func executeRefresh(ctx context.Context, db *pgxpool.Pool, sql string, refresh *Refresh) {
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatalf("Erreur à l'ouverture de la transaction pour le refresh des vues : %s", err.Error())
	}
	for _, current := range strings.Split(sql, ";\n") {
		log.Printf("Refresh - %s - Exécute la requête : '%s'", refresh, current)
		refresh.Run(current)
		_, err = tx.Exec(ctx, current)
		if err != nil {
			refresh.Fail(err.Error())
			err := tx.Rollback(ctx)
			if err != nil {
				log.Fatalf("Erreur lors du rollback de transaction. Cause : %s", err.Error())
			}
			log.Fatalf("Erreur lors de l'exécution de la requête '%s'. Cause : %s", current, err.Error())
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		refresh.Fail(err.Error())
		log.Fatalf("Erreur lors du commit de transaction : %s", err.Error())
	}
	refresh.Finish()
}
