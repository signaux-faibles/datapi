// Package test contient le code commun aux tests
package test

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"path/filepath"
	"time"

	"github.com/jaswdr/faker"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

type dockerContext struct {
	pool           *dockertest.Pool
	containerNames map[containerName]string
}

type containerName string

const datapiDBName containerName = "datapiDBName"
const wekanDBName containerName = "wekanDBName"
const datapiLogDBName containerName = "datapiLogsDBName"

var d *dockerContext

var fake faker.Faker

func init() {
	fake = faker.New()
	name := fake.Internet().User()
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Panicf("Could not connect to docker: %s", err)
	}
	newContext := dockerContext{
		pool: pool,
		containerNames: map[containerName]string{
			datapiDBName:    "datapidb-ti-" + name,
			datapiLogDBName: "datapilogdb-ti-" + name,
			wekanDBName:     "mongodb-ti-" + name,
		},
	}
	d = &newContext
}

// GetDatapiDBURL démarre si nécessaire un container postgres et retourne l'URL pour y accéder
func GetDatapiDBURL() string {
	datapiDB, found := getContainer(datapiDBName)
	if !found {
		datapiDB = startDatapiDB()
	} else {
		log.Printf("le container %s a bien été retrouvé", datapiDB.Container.Name)
	}
	hostAndPort := datapiDB.GetHostPort("5432/tcp")
	dbURL := fmt.Sprintf("postgres://postgres:test@%s/%s?sslmode=disable", hostAndPort, DatapiDatabaseName)
	return dbURL
}

// GetDatapiLogDBURL démarre si nécessaire un container postgres et retourne l'URL pour y accéder
func GetDatapiLogDBURL() string {
	datapiLogDB, found := getContainer(datapiLogDBName)
	if !found {
		datapiLogDB = startDatapiLogDB()
	} else {
		log.Printf("le container %s a bien été retrouvé", datapiLogDB.Container.Name)
	}
	hostAndPort := datapiLogDB.GetHostPort("5432/tcp")
	dbURL := fmt.Sprintf("postgres://postgres:test@%s/%s?sslmode=disable", hostAndPort, DatapiDatabaseName+"_log")
	return dbURL
}

// GetWekanDBURL démarre si nécessaire un container Mongo et retourne l'URL pour y accéder
func GetWekanDBURL() string {
	wekanDB, found := getContainer(wekanDBName)
	if !found {
		wekanDB = startWekanDB()
	}
	hostAndPort := wekanDB.GetHostPort("27017/tcp")
	url := fmt.Sprintf("mongodb://mgo:test@%s/", hostAndPort)
	return url
}

func startDatapiDB() *dockertest.Resource {
	// configuration file for postgres
	postgresConfig, err := filepath.Abs("test/postgresql.conf")
	if err != nil {
		log.Panicf("Could not get absolute path: %s", err)
	}
	// sql dump to currentContext postgres data
	sqlDump, _ := filepath.Abs("test/data")
	// pulls an image, creates a container based on it and runs it
	datapiContainerName := d.containerNames[datapiDBName]
	log.Println("démarre le container datapi-db avec le nom : ", datapiContainerName)

	datapiDB, err := currentPool().RunWithOptions(&dockertest.RunOptions{
		Name:       datapiContainerName,
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=test",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=" + DatapiDatabaseName,
			"listen_addresses = '*'"},
		Mounts: []string{
			postgresConfig + ":/etc/postgresql/postgresql.conf",
			sqlDump + ":/docker-entrypoint-initdb.d",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		killContainer(datapiDB)
		log.Fatal("Could not start datapi_db", err)
	}
	// container stops after 20'
	if err = datapiDB.Expire(600); err != nil {
		killContainer(datapiDB)
		log.Fatal("Could not set expiration on container datapi_db", err)
	}
	return datapiDB
}

func startWekanDB() *dockertest.Resource {
	// pulls an image, creates a container based on it and runs it
	mongoContainerName := d.containerNames[wekanDBName]
	log.Println("démarre le container", mongoContainerName)

	wekanDB, err := currentPool().RunWithOptions(&dockertest.RunOptions{
		Name:       mongoContainerName,
		Repository: "mongo",
		Tag:        "4.0-xenial",
		Env: []string{
			"MONGO_INITDB_ROOT_PASSWORD=test",
			"MONGO_INITDB_ROOT_USERNAME=mgo",
			"MONGO_INITDB_DATABASE=test",
			"listen_addresses = '*'"},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		killContainer(wekanDB)
		log.Fatal("erreur pendant le démarrage du container", mongoContainerName, err)
	}
	// container stops after 60 seconds
	if err = wekanDB.Expire(120); err != nil {
		killContainer(wekanDB)
		log.Fatal("erreur pendant la mise en expiration du container", mongoContainerName, err)
	}
	return wekanDB
}

func startDatapiLogDB() *dockertest.Resource {
	// configuration file for postgres
	postgresConfig, err := filepath.Abs("test/postgresql.conf")
	if err != nil {
		log.Panicf("Could not get absolute path: %s", err)
	}
	// sql dump to currentContext postgres data
	// sqlDump, _ := filepath.Abs("test/data")
	// pulls an image, creates a container based on it and runs it
	datapiLogContainerName := d.containerNames[datapiLogDBName]
	log.Println("démarre le container datapi-log-db avec le nom : ", datapiLogContainerName)

	datapiLogDB, err := currentPool().RunWithOptions(&dockertest.RunOptions{
		Name:       datapiLogContainerName,
		Repository: "postgres",
		Tag:        "15-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=test",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=" + DatapiDatabaseName + "_log",
			"listen_addresses = '*'"},
		Mounts: []string{
			postgresConfig + ":/etc/postgresql/postgresql.conf",
			//sqlDump + ":/docker-entrypoint-initdb.d",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		killContainer(datapiLogDB)
		log.Fatal("Could not start datapi-log-db", err)
	}
	// container stops after 20'
	if err = datapiLogDB.Expire(600); err != nil {
		killContainer(datapiLogDB)
		log.Fatal("Could not set expiration on container datapi-log-db", err)
	}
	return datapiLogDB
}

func killContainer(resource *dockertest.Resource) {
	if resource == nil {
		return
	}
	if err := resource.Close(); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

// Wait4PostgresIsReady attends que le container postgres pour DatapiDb soit prêt
// TODO je pense qu'on peut améliorer cette partie pour la rendre plus agréable à utiliser
func Wait4PostgresIsReady(datapiDBUrl string) {
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool := currentPool()
	pool.MaxWait = 120 * time.Second
	if err := pool.Retry(func() error {
		pg, err := sql.Open("postgres", datapiDBUrl)
		if err != nil {
			return err
		}
		return pg.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
}

func getContainer(name containerName) (*dockertest.Resource, bool) {
	return d.pool.ContainerByName(d.containerNames[name])
}

func currentPool() *dockertest.Pool {
	return d.pool
}

func GenerateRandomPort() string {
	n, err := rand.Int(rand.Reader, big.NewInt(500))
	n.Add(n, big.NewInt(30000))
	if err != nil {
		fmt.Println("erreur pendant la génération d'un nombre aléatoire : ", err)
		return ""
	}
	return n.String()
}
