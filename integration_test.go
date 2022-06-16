package datapi

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/signaux-faibles/datapi/src/core"
	test "github.com/signaux-faibles/datapi/src/test"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
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
		killContainer(datapiDb)
		log.Fatal("Could not start datapi_db", err)
	}
	// container stops after 60 seconds
	if err = datapiDb.Expire(120); err != nil {
		killContainer(datapiDb)
		log.Fatal("Could not set expiration on container datapi_db", err)
	}
	hostAndPort := datapiDb.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://postgres:test@%s/datapi_test?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url: ", databaseUrl)

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
	core.StartDatapi()
	go core.RunAPI()
	time.Sleep(1 * time.Second) // time to expose API

	//Run tests
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(datapiDb); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
	os.Exit(code)
}

func TestListes(t *testing.T) {
	_, indented, _ := test.Get(t, "/listes")
	test.ProcessGoldenFile(t, "test/data/listes.json.gz", indented)
}

func TestFollow(t *testing.T) {
	test.RazEtablissementFollowing(t)
	sirets := test.SelectSomeSiretsToFollow(t)

	params := map[string]interface{}{
		"comment":  "test",
		"category": "test",
	}

	for _, siret := range sirets {
		t.Logf("suivi de l'établissement %s", siret)
		resp, _, _ := test.Post(t, "/follow/"+siret, params)
		if resp.StatusCode != 201 {
			t.Errorf("le suivi a échoué: %d", resp.StatusCode)
		}
	}

	for _, siret := range sirets {
		t.Logf("suivi doublon de l'établissement %s", siret)
		resp, _, _ := test.Post(t, "/follow/"+siret, params)
		if resp.StatusCode != 204 {
			t.Errorf("le doublon n'a pas été détecté correctement: %d", resp.StatusCode)
		}
	}

	t.Log("Vérification de /follow")
	core.Db().Exec(context.Background(), "update etablissement_follow set since='2020-03-01'")
	_, indented, _ := test.Get(t, "/follow")
	test.ProcessGoldenFile(t, "test/data/follow.json.gz", indented)
}

func TestSearch(t *testing.T) {
	test.RazEtablissementFollowing(t)
	for _, siret := range test.SelectSomeSiretsToFollow(t) {
		test.FollowEtablissement(t, siret)
	}

	// tester le retour 400 en cas de recherche trop courte
	t.Log("/etablissement/search retourne 400")
	params := map[string]interface{}{
		"search": "t",
	}
	resp, _, _ := test.Post(t, "/etablissement/search", params)
	if resp.StatusCode != 400 {
		t.Errorf("mauvais status retourné: %d", resp.StatusCode)
	}

	// tester la recherche par chaine de caractères
	rows, err := core.Db().Query(context.Background(), `select distinct substring(e.siret from 1 for 3) from etablissement e
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
		_, indented, _ := test.Post(t, "/etablissement/search", params)
		goldenFilePath := fmt.Sprintf("test/data/search-%d.json.gz", i)
		test.ProcessGoldenFile(t, goldenFilePath, indented)
		i++
	}

	// tester par département
	var departements []string
	var siret string
	err = core.Db().QueryRow(
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
	_, indented, _ := test.Post(t, "/etablissement/search", params)
	goldenFilePath := "test/data/searchDepartement.json.gz"
	test.ProcessGoldenFile(t, goldenFilePath, indented)
	i++

	// tester par activité
	err = core.Db().QueryRow(
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
	_, indented, _ = test.Post(t, "/etablissement/search", params)
	goldenFilePath = "test/data/searchActivites.json.gz"
	test.ProcessGoldenFile(t, goldenFilePath, indented)
}

// TestPGE ; cette fonction teste si le PGE est bien retourné par l'API si
// - l'entreprise a un pge actif
// - l'utilisateur connecté a bien les habilitations
//
// pour que ce test fonctionne il faut vérifier que le fichier dump
// des données de test contient bien les informations suivantes ;
// 020523337	true
// 221129065	true
// 336400422	false
func TestPGE(t *testing.T) {
	assertions := assert.New(t)

	test.RazEtablissementFollowing(t)

	tests := []test.PgeTest{
		// pge is true in entreprise_pge and has habilitation => true
		{Siren: "020523337", HasPGE: &core.True, MustFollow: core.True, ExpectedPGE: &core.True},
		// pge is true in entreprise_pge and has no habilitation => nil
		{Siren: "221129065", HasPGE: &core.True, MustFollow: core.False, ExpectedPGE: nil},
		// pge is false in  entreprise_pge and has no habilitation => nil
		{Siren: "336400422", HasPGE: &core.False, MustFollow: core.False, ExpectedPGE: nil},
		// pge is false in  entreprise_pge and has habilitation => false
		{Siren: "386322594", HasPGE: &core.False, MustFollow: core.True, ExpectedPGE: &core.False},
		// not exists in entreprise_pge and has habilitation => nil
		{Siren: "417193286", HasPGE: nil, MustFollow: core.True, ExpectedPGE: nil},
	}
	test.InsertPGE(t, tests)

	for _, pgeTest := range tests {
		t.Logf("Test PGE for entreprise '%s'", pgeTest.Siren)
		siren := pgeTest.Siren

		if pgeTest.MustFollow {
			test.FollowEntreprise(t, siren)
		}

		resp, data, _ := test.Get(t, "/entreprise/get/"+siren)
		assertions.Equal(200, resp.StatusCode)
		actual := test.JsonToEntreprise(t, data)
		assertions.Equal(pgeTest.ExpectedPGE, actual.PGEActif)
	}
}

func loadTestConfig(postgresURL string) {
	log.Println("loading test config")
	core.LoadConfig("test", "config", "migrations")
	viper.Set("postgres", postgresURL)
	datapiPort := strconv.Itoa(rand.Intn(500) + 30000)
	viper.Set("bind", ":"+datapiPort)
	os.Setenv("DATAPI_URL", "http://localhost:"+datapiPort)
}

func killContainer(resource *dockertest.Resource) {
	if resource == nil {
		return
	}
	if err := resource.Close(); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}
