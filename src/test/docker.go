package test

import (
	"database/sql"
	"fmt"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log"
	"path/filepath"
	"strconv"
	"time"
)

func StartDatapiDBAndGetURL(pool *dockertest.Pool) (*dockertest.Resource, string) {
	// configuration file for postgres
	postgresConfig, err := filepath.Abs("test/postgresql.conf")
	if err != nil {
		log.Panicf("Could not get absolute path: %s", err)
	}
	// sql dump to initialize postgres data
	sqlDump, err := filepath.Abs("test/data")

	// pulls an image, creates a container based on it and runs it
	datapiContainerName := "datapidb-ti-" + strconv.Itoa(time.Now().Nanosecond())
	log.Println("trying start datapi-db")

	datapiDb, err := pool.RunWithOptions(&dockertest.RunOptions{
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
	hostAndPort := datapiDb.GetHostPort("5432/tcp")
	dbURL := fmt.Sprintf("postgres://postgres:test@%s/datapi_test?sslmode=disable", hostAndPort)
	return datapiDb, dbURL
}

func StartWekanDBAndGetURL(pool *dockertest.Pool) (*dockertest.Resource, string) {
	// pulls an image, creates a container based on it and runs it
	mongoContainerName := "mongodb-ti-" + strconv.Itoa(time.Now().Nanosecond())
	log.Println("trying start mongo-db")

	wekanDb, err := pool.RunWithOptions(&dockertest.RunOptions{
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
	hostAndPort := wekanDb.GetHostPort("27017/tcp")
	url := fmt.Sprintf("mongodb://mgo:test@%s/", hostAndPort)
	return wekanDb, url
}

func killContainer(resource *dockertest.Resource) {
	if resource == nil {
		return
	}
	if err := resource.Close(); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func Wait4DatapiDb(pool *dockertest.Pool, datapiDBUrl string) {
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
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
