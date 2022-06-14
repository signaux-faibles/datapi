package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/spf13/viper"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// TestMain : lance datapi ainsi qu'un conteneur postgres bien paramétrer
// les informations de base de données doivent être identique dans :
// - le fichier de configuration de test -> test/config.toml
// - le fichier de création et d'import de données dans la base -> test/data/testData.sql.gz
// - la configuration du container
func TestMain(m *testing.M) {

	rand.Seed(time.Now().UnixNano())

	// configuration file for postgres
	postgresConfig, err := filepath.Abs("test/postgresql.conf")
	if err != nil {
		log.Panicf("Could not get absolute path: %s", err)
	}
	// sql dump to initialize postgres data
	sqlDump, err := filepath.Abs("test/data")

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Panicf("Could not connect to docker: %s", err)
	}
	// pulls an image, creates a container based on it and runs it
	datapiContainerName := "datapidb-ti-" + strconv.Itoa(time.Now().Nanosecond())
	log.Println("trying start datapi-db")

	datapiDb, err := pool.RunWithOptions(&dockertest.RunOptions{
		Name:       datapiContainerName,
		Repository: "postgres",
		Tag:        "10-alpine",
		//Tag:    "14-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=test",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=datapi_test",
			"listen_addresses = '*'"},
		Mounts: []string{
			postgresConfig + ":/etc/postgresql/postgresql.conf",
			sqlDump + ":/docker-entrypoint-initdb.d",
		},
	}, func(config *docker.HostConfig) {
		//set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		kill(datapiDb)
		log.Fatal("Could not start datapi_db", err)
	}
	// container stops after 60 seconds
	if err = datapiDb.Expire(120); err != nil {
		kill(datapiDb)
		log.Fatal("Could not set expiration on container datapi_db", err)
	}
	hostAndPort := datapiDb.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://postgres:test@%s/datapi_test?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url: ", databaseUrl)
	datapiDb.Expire(120) // Tell docker to hard kill the container in 120 seconds

	loadTestConfig(databaseUrl)

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		db, err := sql.Open("postgres", databaseUrl)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	// run datapi
	startDatapi()
	go runAPI()
	time.Sleep(2 * time.Second)

	//Run tests
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(datapiDb); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	os.Exit(code)
}

func loadTestConfig(postgresURL string) {
	log.Println("loading test config")
	loadConfig("test", "config", "migrations")
	viper.Set("postgres", postgresURL)
	datapiPort := strconv.Itoa(rand.Intn(500) + 30000)
	viper.Set("bind", ":"+datapiPort)
	os.Setenv("DATAPI_URL", "http://localhost:"+datapiPort)
}

func kill(resource *dockertest.Resource) {
	if resource == nil {
		return
	}
	if err := resource.Close(); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func TestListes(t *testing.T) {
	_, indented, _ := get(t, "/listes")
	processGoldenFile(t, "test/data/listes.json.gz", indented)
}

func TestSearch(t *testing.T) {
	// tester le retour 400 en cas de recherche trop courte
	t.Log("/etablissement/search retourne 400")
	params := map[string]interface{}{
		"search": "t",
	}
	resp, _, _ := post(t, "/etablissement/search", params)
	if resp.StatusCode != 400 {
		t.Errorf("mauvais status retourné: %d", resp.StatusCode)
	}

	// tester la recherche par chaine de caractères
	rows, err := db.Query(context.Background(), `select distinct substring(e.siret from 1 for 3) from etablissement e
	inner join departements d on d.code = e.departement
	inner join regions r on r.id = d.id_region
	where r.libelle in ('Bourgogne-Franche-Comté', 'Auvergne-Rhône-Alpes')
	order by substring(e.siret from 1 for 3)
	limit 10
	`)
	if err != nil {
		t.Errorf("impossible de se connecter à la base: %s", err.Error())
	}

	i := 0
	for rows.Next() {
		var siret string
		err := rows.Scan(&siret)
		if err != nil {
			t.Errorf("siret illisible: %s", err.Error())
		}

		params := make(map[string]interface{})
		params["search"] = siret
		params["ignoreZone"] = false
		params["ignoreRoles"] = false
		t.Logf("la recherche %s est bien de la forme attendue", siret)
		_, indented, _ := post(t, "/etablissement/search", params)
		goldenFilePath := fmt.Sprintf("test/data/search-%d.json.gz", i)
		processGoldenFile(t, goldenFilePath, indented)
		i++
	}

	// tester par département
	var departements []string
	var siret string
	err = db.QueryRow(
		context.Background(),
		`select array_agg(distinct departement), substring(first(siret) from 1 for 3) from etablissement where departement < '10' and departement != '00'`,
	).Scan(&departements, &siret)
	if err != nil {
		t.Errorf("impossible de se connecter à la base: %s", err.Error())
	}

	params = map[string]interface{}{
		"departements": departements,
		"search":       siret,
		"ignoreZone":   true,
		"ignoreRoles":  true,
	}

	t.Log("la recherche filtrée par départements est bien de la forme attendue")
	_, indented, _ := post(t, "/etablissement/search", params)
	goldenFilePath := "test/data/searchDepartement.json.gz"
	processGoldenFile(t, goldenFilePath, indented)
	i++

	// tester par activité
	err = db.QueryRow(
		context.Background(),
		`select substring(first(siret) from 1 for 3)
		 from etablissement e
		 inner join v_naf n on n.code_n5 = e.code_activite
		 where code_n1 in ('A', 'B', 'C')
		`,
	).Scan(&siret)
	if err != nil {
		t.Errorf("impossible de se connecter à la base: %s", err.Error())
	}

	params = map[string]interface{}{
		"activites":   []string{"A", "B", "C"},
		"search":      siret,
		"ignoreZone":  true,
		"ignoreRoles": true,
	}

	t.Log("la recherche filtrée par activites est bien de la forme attendue")
	_, indented, _ = post(t, "/etablissement/search", params)
	goldenFilePath = "test/data/searchActivites.json.gz"
	processGoldenFile(t, goldenFilePath, indented)
}
