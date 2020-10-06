package main

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"sync"

	pgx "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/spf13/viper"

	_ "github.com/lib/pq"
)

type migrationScript struct {
	fileName string
	content  []byte
	hash     string
}

func connectDB() *pgxpool.Pool {
	pgConnStr := viper.GetString("postgres")
	db, err := pgxpool.Connect(context.Background(), pgConnStr)
	if err != nil {
		log.Fatal("database connexion:" + err.Error())
	}

	log.Print("connected to postgresql database")
	dbMigrations := listDatabaseMigrations(db)
	dirMigrations := listDirMigrations()
	updateMigrations := compareMigrations(dbMigrations, dirMigrations)
	log.Printf("%d embedded migrations, %d db migrations", len(dirMigrations), len(dbMigrations))
	if len(dbMigrations) < len(dirMigrations) {
		log.Printf("running %d new migrations", len(updateMigrations))
		runMigrations(updateMigrations, db)
	}

	ref = loadReferences(db)

	return db
}

func listDatabaseMigrations(db *pgxpool.Pool) []migrationScript {
	var exists bool
	err := db.QueryRow(context.Background(),
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

func runMigrations(migrationScripts []migrationScript, db *pgxpool.Pool) {
	tx, err := db.Begin(context.Background())

	if err != nil {
		panic("something bad is happening with database" + err.Error())
	}
	for _, m := range migrationScripts {
		_, err = tx.Exec(context.Background(), string(m.content))
		if err != nil {
			tx.Rollback(context.Background())
			panic("error migrating " + m.fileName + ", no changes commited. see details:\n" + err.Error())
		}
		_, err = tx.Exec(
			context.Background(),
			"insert into migrations (filename, hash) values ($1, $2)", m.fileName, string(m.hash[:]))
		if err != nil {
			tx.Rollback(context.Background())
			panic("error inserting " + m.fileName + " in migration table, no changes commited. see details:\n" + err.Error())
		}
		log.Printf("%s rolled, registered with hash %s", m.fileName, m.hash)
	}

	tx.Commit(context.Background())
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

func newBatchRunner(tx *pgx.Tx) (chan pgx.Batch, *sync.WaitGroup) {
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

type reference struct {
	zones map[string][]string
}

func loadReferences(db *pgxpool.Pool) reference {
	sqlZones := `select r.libelle as region, d.code as departement 
	from regions r 
	inner join departements d on d.id_region = r.id 
	union select 'France entière', code from departements
	order by region, departement`
	rows, err := db.Query(context.Background(), sqlZones)
	if err != nil {
		panic("can't load references: " + err.Error())
	}

	var ref = make(map[string][]string)

	for rows.Next() {
		region, departement := "", ""
		err := rows.Scan(&region, &departement)
		if err != nil {
			panic("can't load references: " + err.Error())
		}

		ref[departement] = []string{departement}
		if _, ok := ref[region]; !ok {
			ref[region] = nil
		}
		ref[region] = append(ref[region], departement)
	}

	result := reference{
		zones: ref,
	}

	return result
}

func createIndexesBatch() pgx.Batch {
	batch := pgx.Batch{}
	batch.Queue("create index idx_etablissement_siret on etablissement (siret) where version = 0;")
	batch.Queue("create index idx_etablissement_siren on etablissement (siren) where version = 0;")
	batch.Queue("create index idx_etablissement_departement on etablissement (departement) where version = 0;")
	batch.Queue("create index idx_etablissement_hash on etablissement (siret, hash) where version in (0, 1);")
	batch.Queue("create index idx_etablissement_naf on etablissement (ape) where version = 0;")
	batch.Queue("create index idx_entreprise_siren on entreprise (siren) where version = 0;")
	batch.Queue("create index idx_entreprise_hash on entreprise (siren, hash) where version in (0, 1);")
	batch.Queue("create index idx_etablissement_procol_siret on etablissement_procol (siret) where version = 0;")
	batch.Queue("create index idx_etablissement_procol_hash on etablissement_procol (siret, hash) where version in (0, 1);")
	batch.Queue("create index idx_etablissement_delai_siret on etablissement_delai (siret) where version = 0;")
	batch.Queue("create index idx_etablissement_delai_hash on etablissement_delai (siret, hash) where version in (0, 1);")
	batch.Queue("create index idx_etablissement_periode_urssaf_siret on etablissement_periode_urssaf (siret) where version = 0;")
	batch.Queue("create index idx_etablissement_periode_urssaf_siren on etablissement_periode_urssaf (siren) where version = 0;")
	batch.Queue("create index idx_etablissement_periode_urssaf_hash on etablissement_periode_urssaf (siret, hash) where version in (0, 1);")
	batch.Queue("create index idx_etablissement_apconso_siret on etablissement_apconso (siret) where version = 0;")
	batch.Queue("create index idx_etablissement_apconso_siren on etablissement_apconso (siren) where version = 0;")
	batch.Queue("create index idx_etablissement_apconso_hash on etablissement_apconso (siret, hash) where version in (0, 1);")
	batch.Queue("create index idx_etablissement_apdemande_siret on etablissement_apdemande (siret) where version = 0;")
	batch.Queue("create index idx_etablissement_apdemande_siren on etablissement_apdemande (siren) where version = 0;")
	batch.Queue("create index idx_etablissement_apdemande_hash on etablissement_apdemande (siret, hash) where version in (0, 1);")
	batch.Queue("create index idx_score_siret on score (siret) where version = 0;")
	batch.Queue("create index idx_score_siren on score (siren) where version = 0;")
	batch.Queue("create index idx_score_alert on score (alert) where version = 0;")
	batch.Queue("create index idx_score_liste on score (libelle_liste) where version = 0;")
	batch.Queue("create index idx_score_score on score (score) where version = 0;")
	batch.Queue("create index idx_score_hash on score (siret, hash) where version in (0, 1);")
	batch.Queue("create index idx_liste_libelle on liste (libelle) where version = 0;")
	batch.Queue("create index idx_naf_code_niveau on naf (code, niveau);")
	batch.Queue("create index idx_etablissement_follow on etablissement_follow (siret, username, active);")
	batch.Queue("create index idx_naf_code on naf  (code);")
	batch.Queue("create index idx_naf_n1 on naf (id_n1);")
	return batch
}

func dropIndexesBatch() pgx.Batch {
	batch := pgx.Batch{}
	batch.Queue("drop index idx_etablissement_siret;")
	batch.Queue("drop index idx_etablissement_siren;")
	batch.Queue("drop index idx_etablissement_departement;")
	batch.Queue("drop index idx_etablissement_hash;")
	batch.Queue("drop index idx_etablissement_naf;")
	batch.Queue("drop index idx_entreprise_siren;")
	batch.Queue("drop index idx_entreprise_hash;")
	batch.Queue("drop index idx_etablissement_procol_siret;")
	batch.Queue("drop index idx_etablissement_procol_hash;")
	batch.Queue("drop index idx_etablissement_delai_siret;")
	batch.Queue("drop index idx_etablissement_delai_hash;")
	batch.Queue("drop index idx_etablissement_periode_urssaf_siret;")
	batch.Queue("drop index idx_etablissement_periode_urssaf_siren;")
	batch.Queue("drop index idx_etablissement_periode_urssaf_hash;")
	batch.Queue("drop index idx_etablissement_apconso_siret;")
	batch.Queue("drop index idx_etablissement_apconso_siren;")
	batch.Queue("drop index idx_etablissement_apconso_hash;")
	batch.Queue("drop index idx_etablissement_apdemande_siret;")
	batch.Queue("drop index idx_etablissement_apdemande_siren;")
	batch.Queue("drop index idx_etablissement_apdemande_hash;")
	batch.Queue("drop index idx_score_siret;")
	batch.Queue("drop index idx_score_siren;")
	batch.Queue("drop index idx_score_alert;")
	batch.Queue("drop index idx_score_liste;")
	batch.Queue("drop index idx_score_score;")
	batch.Queue("drop index idx_score_hash;")
	batch.Queue("drop index idx_liste_libelle;")
	batch.Queue("drop index idx_naf_code_niveau;")
	batch.Queue("drop index idx_etablissement_follow;")
	batch.Queue("drop index idx_naf_code;")
	batch.Queue("drop index idx_naf_n1;")
	return batch
}
