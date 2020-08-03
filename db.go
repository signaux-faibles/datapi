package main

import (
	"crypto/sha1"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"sort"

	"github.com/spf13/viper"

	_ "github.com/lib/pq"
)

type migrationScript struct {
	fileName string
	content  []byte
	hash     string
}

func connectDB() *sql.DB {
	pgConnStr := viper.GetString("postgres")
	db, err := sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Fatal("database connexion:" + err.Error())
	}
	log.Print("Connected to postgresql database")
	dbMigrations := listDatabaseMigrations(db)
	dirMigrations := listDirMigrations()
	updateMigrations := compareMigrations(dbMigrations, dirMigrations)
	log.Printf("%d embedded migrations, %d db migrations", len(dirMigrations), len(dbMigrations))
	if len(dbMigrations) < len(dirMigrations) {
		log.Printf("running %d new migrations", len(updateMigrations))
		runMigrations(updateMigrations, db)
	}

	return db
}

func listDatabaseMigrations(db *sql.DB) []migrationScript {
	var exists bool
	row := db.QueryRow(
		`SELECT EXISTS 
		(SELECT 1
		 FROM information_schema.tables 
		 WHERE table_schema = 'public'
	   AND table_name = 'migrations');`,
	)
	row.Scan(&exists)
	if !exists {
		return nil
	}
	dbMigrationsCursor, err := db.Query(`
		select filename, hash 
		from migrations
		order by filename`)
	if err != nil {
		panic("can't get migrations from database: " + err.Error())
	}
	var dbMigrations []migrationScript
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
	files, err := ioutil.ReadDir(viper.GetString("migrationsDir"))
	if err != nil {
		panic("migrations not found: " + err.Error())
	}
	var dirMigrations []migrationScript
	hasher := sha1.New()
	for _, f := range files {
		content, err := ioutil.ReadFile("./migrations/" + f.Name())
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

func runMigrations(migrationScripts []migrationScript, db *sql.DB) {
	tx, err := db.Begin()

	if err != nil {
		panic("something bad is happening with database" + err.Error())
	}
	for _, m := range migrationScripts {
		_, err = tx.Exec(string(m.content))
		if err != nil {
			tx.Rollback()
			panic("error migrating " + m.fileName + ", no changes commited. see details:\n" + err.Error())
		}
		_, err = tx.Exec("insert into migrations (filename, hash) values ($1, $2)", m.fileName, string(m.hash[:]))
		if err != nil {
			tx.Rollback()
			panic("error inserting " + m.fileName + " in migration table, no changes commited. see details:\n" + err.Error())
		}
		log.Printf("%s rolled, registered with hash %s", m.fileName, m.hash)
	}

	tx.Commit()
}
