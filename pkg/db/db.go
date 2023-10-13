// Package db : contient les méthodes qui concernent la base de données.
package db

import (
	"context"
	"crypto/sha1"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
)

var db *pgxpool.Pool

type migrationScript struct {
	fileName string
	content  []byte
	hash     string
}

// Get : expose le pool de connexion Datapi
func Get() *pgxpool.Pool {
	if db == nil {
		Init()
	}
	return db
}

// Init : se connecte à la base de données et exécute les scripts de migrations nécessaires
func Init() {
	pgConnStr := viper.GetString("postgres")
	pool, err := pgxpool.New(context.Background(), pgConnStr)
	if err != nil {
		slog.Error("erreur lors de la connexion à la base", slog.Any("error", err))
		panic(err)
	}

	slog.Info("connected to postgresql database")
	dbMigrations := listDatabaseMigrations(pool)
	dirMigrations := listDirMigrations()
	updateMigrations := compareMigrations(dbMigrations, dirMigrations)
	slog.Info(
		"liste les migrations",
		slog.Int("embedded", len(dirMigrations)),
		slog.Int("db", len(dbMigrations)),
	)
	if len(dbMigrations) < len(dirMigrations) {
		slog.Info("démarre les nouvelles migrations", slog.Any("new", len(updateMigrations)))
		runMigrations(updateMigrations, pool)
	}

	db = pool
}

func listDatabaseMigrations(db *pgxpool.Pool) []migrationScript {
	var exists bool
	db.QueryRow(context.Background(),
		`select exists
		(select 1
		 from information_schema.tables
		 where table_schema = 'public'
	   and table_name = 'migrations');`,
	).Scan(&exists)
	if !exists {
		return nil
	}
	dbMigrationsCursor, err := db.Query(context.Background(),
		`select filename, hash
		from migrations
		order by filename`)
	if err != nil {
		panic("can't get migrations from database: " + err.Error())
	}
	var dbMigrations []migrationScript
	defer dbMigrationsCursor.Close()

	for dbMigrationsCursor.Next() {
		var m migrationScript
		dbMigrationsCursor.Scan(&m.fileName, &m.hash)
		dbMigrations = append(dbMigrations, m)
	}
	return dbMigrations
}

func compareMigrations(db []migrationScript, dir []migrationScript) []migrationScript {
	if len(dir) < len(db) {
		panic("database contains more migrations than code does, you probably run an old version, consider updating code.")
	}
	for l, m := range db {
		if m.fileName != dir[l].fileName {
			panic(m.fileName + " != " + dir[l].fileName + ". unexpected migration file name.")
		}
		if m.hash != dir[l].hash {
			panic("migration " + m.fileName + " is not the same in database and disk, something nasty is happening, you definitely need to troubleshoot this.")
		}
	}
	return dir[len(db):]
}

func listDirMigrations() []migrationScript {
	migrationsDir := viper.GetString("migrationsDir")
	wd, err := os.Getwd()
	if err != nil {
		log.Panicf(
			"erreur lors de la récupération du répertoire de travail : %s",
			err.Error(),
		)
	}
	if !filepath.IsAbs(migrationsDir) {
		migrationsDir, err = filepath.Abs(filepath.Join(wd, migrationsDir))
		if err != nil {
			log.Panicf(
				"erreur lors de la récupération du chemin absolu du répertoire de migration '%s' : %s",
				migrationsDir,
				err.Error(),
			)
		}
	}
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		slog.Error("le répertoire des fichiers de migration est introuvable",
			slog.String("migrationsDir", migrationsDir),
			slog.String("wd", wd),
			slog.Any("error", err),
		)
		panic(err)
	}
	var dirMigrations []migrationScript
	hasher := sha1.New()
	for _, f := range files {
		content, err := os.ReadFile(filepath.Join(migrationsDir, f.Name()))
		if err != nil {
			panic("error reading migration files: " + err.Error())
		}
		hasher.Write(content)

		m := migrationScript{
			fileName: f.Name(),
			content:  content,
			hash:     fmt.Sprintf("%x", hasher.Sum(nil)),
		}
		dirMigrations = append(dirMigrations, m)
	}
	sort.Slice(dirMigrations, func(i, j int) bool {
		return dirMigrations[i].fileName < dirMigrations[j].fileName
	})
	return dirMigrations
}

func runMigrations(migrationScripts []migrationScript, db *pgxpool.Pool) {
	ctx := context.Background()
	tx, err := db.Begin(ctx)

	if err != nil {
		panic("something bad is happening with database" + err.Error())
	}
	// Defer a rollback in case anything fails.
	defer tx.Rollback(ctx)
	for _, m := range migrationScripts {
		_, err = tx.Exec(ctx, string(m.content))
		if err != nil {
			slog.Error("erreur pendant la migration. Pas de changement effectué", slog.Any("error", err), slog.String("filename", m.fileName), slog.Any("hash", m.hash))
			panic(err)
		}
		_, err = tx.Exec(
			ctx,
			"insert into migrations (filename, hash) values ($1, $2)", m.fileName, m.hash[:])
		if err != nil {
			panic("error inserting " + m.fileName + " in migration table, no changes commited. see details:\n" + err.Error())
		}
		slog.Info("migration effectuée", slog.String("filename", m.fileName), slog.Any("hash", m.hash))
	}
	err = tx.Commit(ctx)
	if err != nil {
		slog.Error("erreur pendant le commit de transaction", slog.Any("error", err))
		panic(err)
	}
}

func runBatch(tx *pgx.Tx, batch *pgx.Batch) {
	r := (*tx).SendBatch(context.Background(), batch)
	err := r.Close()
	if err != nil {
		// TODO: meilleur usage des contextes pour remonter l'erreur et annuler la transaction
		fmt.Println("Emergency rollback:" + err.Error())
		(*tx).Rollback(context.Background())
	}
}

// NewBatchRunner : crée un channel d'exécution de Batch
func NewBatchRunner(tx *pgx.Tx) (chan pgx.Batch, *sync.WaitGroup) {
	var batches = make(chan pgx.Batch)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for batch := range batches {
			runBatch(tx, &batch)
		}
		wg.Done()
	}()

	return batches, &wg
}
