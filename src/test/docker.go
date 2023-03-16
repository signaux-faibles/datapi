// Package test contient le code commun aux tests
package test

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log"
	"path/filepath"
	"sync/atomic"
	"time"
)

type dockerContext struct {
	pool           *dockertest.Pool
	containerNames map[containerName]string
}

type containerName string

const (
	datapiDbName containerName = "datapiDbName"
	wekanDbName  containerName = "wekanDbName"
)

var current atomic.Value

// GetDatapiDbURL démarre si nécessaire un container postgres et retourne l'URL pour y accéder
func GetDatapiDbURL() string {
	datapiDb, found := getContainer(datapiDbName)
	if !found {
		datapiDb = startDatapiDb()
	}
	hostAndPort := datapiDb.GetHostPort("5432/tcp")
	dbURL := fmt.Sprintf("postgres://postgres:test@%s/datapi_test?sslmode=disable", hostAndPort)
	return dbURL
}

// GetWekanDbURL démarre si nécessaire un container Mongo et retourne l'URL pour y accéder
func GetWekanDbURL() string {
	wekanDb, found := getContainer(wekanDbName)
	if !found {
		wekanDb = startWekanDb()
	}
	hostAndPort := wekanDb.GetHostPort("27017/tcp")
	url := fmt.Sprintf("mongodb://mgo:test@%s/", hostAndPort)
	return url
}

func startDatapiDb() *dockertest.Resource {
	// configuration file for postgres
	postgresConfig, err := filepath.Abs("test/postgresql.conf")
	if err != nil {
		log.Panicf("Could not get absolute path: %s", err)
	}
	// sql dump to currentContext postgres data
	sqlDump, err := filepath.Abs("test/data")
	// pulls an image, creates a container based on it and runs it
	datapiContainerName := currentContext().containerNames[datapiDbName]
	log.Println("trying start datapi-db")

	datapiDb, err := currentPool().RunWithOptions(&dockertest.RunOptions{
		Name:       datapiContainerName,
		Repository: "postgres",
		Tag:        "15-alpine",
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
	// container stops after 20'
	if err = datapiDb.Expire(600); err != nil {
		killContainer(datapiDb)
		log.Fatal("Could not set expiration on container datapi_db", err)
	}
	return datapiDb
}

func startWekanDb() *dockertest.Resource {
	// pulls an image, creates a container based on it and runs it
	mongoContainerName := currentContext().containerNames[wekanDbName]
	log.Println("trying start mongo-db")

	wekanDb, err := currentPool().RunWithOptions(&dockertest.RunOptions{
		Name:       mongoContainerName,
		Repository: "mongo",
		Tag:        "4.0-xenial",
		Env: []string{
			"MONGO_INITDB_ROOT_PASSWORD=test",
			"MONGO_INITDB_ROOT_USERNAME=mgo",
			"MONGO_INITDB_DATABASE=test",
			"listen_addresses = '*'"},
	}, func(config *docker.HostConfig) {
		//set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		killContainer(wekanDb)
		log.Fatal("Could not start datapi_db", err)
	}
	// container stops after 60 seconds
	if err = wekanDb.Expire(120); err != nil {
		killContainer(wekanDb)
		log.Fatal("Could not set expiration on container datapi_db", err)
	}
	return wekanDb
}

func killContainer(resource *dockertest.Resource) {
	if resource == nil {
		return
	}
	if err := resource.Close(); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

// Wait4DatapiDb attends que le container postgres pour DatapiDb soit prêt
// TODO je pense qu'on peut améliorer cette partie pour la rendre plus agréable à utiliser
func Wait4DatapiDb(datapiDBUrl string) {
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

func currentContext() dockerContext {
	contexte := current.Load()
	if contexte != nil {
		return contexte.(dockerContext)
	}
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Panicf("Could not connect to docker: %s", err)
	}
	newContext := dockerContext{
		pool: pool,
		containerNames: map[containerName]string{
			datapiDbName: "datapidb-ti-" + uuid.New().String(),
			wekanDbName:  "mongodb-ti-" + uuid.New().String(),
		},
	}
	current.Store(newContext)
	return newContext
}

func getContainer(name containerName) (*dockertest.Resource, bool) {
	contexte := currentContext()
	return contexte.pool.ContainerByName(contexte.containerNames[name])
}

func currentPool() *dockertest.Pool {
	return currentContext().pool
}
