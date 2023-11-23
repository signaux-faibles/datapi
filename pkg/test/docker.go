// Package test contient le code commun aux tests
package test

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
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

var d *dockerContext

var fake faker.Faker

func init() {
	fake = NewFaker()
	name := fake.Internet().User()
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		slog.Error("Erreur lors de la connexion au pool Docker", slog.Any("error", err))
		panic(err)
	}
	newContext := dockerContext{
		pool: pool,
		containerNames: map[containerName]string{
			datapiDBName: "datapidb-ti-" + name,
			wekanDBName:  "mongodb-ti-" + name,
		},
	}
	slog.Info("crée un nouveau contexte Docker",
		slog.String("datapiDBName", newContext.containerNames[datapiDBName]),
		slog.String("wekanDBName", newContext.containerNames[wekanDBName]),
	)
	d = &newContext
}

// GetDatapiDBURL démarre si nécessaire un container postgres et retourne l'URL pour y accéder
func GetDatapiDBURL() string {
	datapiDB, found := getContainer(datapiDBName)
	if !found {
		datapiDB = startDatapiDB()
	} else {
		slog.Info("le container datapi-db a bien été retrouvé", slog.String("name", datapiDB.Container.Name))
	}
	hostAndPort := datapiDB.GetHostPort("5432/tcp")
	dbURL := fmt.Sprintf("postgres://postgres:test@%s/%s?sslmode=disable", hostAndPort, DatapiDatabaseName)
	wait4PostgresIsReady(dbURL)
	return dbURL
}

// GetDatapiLogsDBURL démarre si nécessaire un container postgres et retourne l'URL pour y accéder
func GetDatapiLogsDBURL() string {
	datapiLogDB, found := getContainer(datapiDBName)
	if !found {
		datapiLogDB = startDatapiDB()
	} else {
		slog.Info("le container datapilog-db a bien été retrouvé", slog.String("name", datapiLogDB.Container.Name))
	}
	hostAndPort := datapiLogDB.GetHostPort("5432/tcp")
	dbURL := fmt.Sprintf("postgres://postgres:test@%s/%s?sslmode=disable", hostAndPort, DatapiLogsDatabaseName)
	wait4PostgresIsReady(dbURL)
	return dbURL
}

// GetWekanDBURL démarre si nécessaire un container Mongo et retourne l'URL pour y accéder
func GetWekanDBURL() string {
	wekanDB, found := getContainer(wekanDBName)
	if !found {
		wekanDB = startWekanDB()
	} else {
		slog.Info("le container wekan-db a bien été retrouvé", slog.String("name", wekanDB.Container.Name))
	}
	hostAndPort := wekanDB.GetHostPort("27017/tcp")
	url := fmt.Sprintf("mongodb://mgo:test@%s/", hostAndPort)
	return url
}

func startDatapiDB() *dockertest.Resource {
	// configuration file for postgres
	postgresConfig := ProjectPathOf("test/postgresql.conf")
	// sql dump to currentContext postgres data
	sqlDump := ProjectPathOf("test/initDB")

	// pulls an image, creates a container based on it and runs it
	datapiContainerName := d.containerNames[datapiDBName]
	slog.Info("démarre le container datapi-db", slog.String("name", datapiContainerName))

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
		slog.Error(
			"erreur pendant le démarrage du container",
			slog.String("container", datapiContainerName),
			slog.Any("error", err),
		)
		panic(err)
	}
	// container stops after 20'
	if err = datapiDB.Expire(600); err != nil {
		killContainer(datapiDB)
		slog.Error(
			"erreur pendant la configuration de l'expiration du container",
			slog.String("container", datapiContainerName),
			slog.Any("error", err),
		)
		panic(err)
	}
	return datapiDB
}

func startWekanDB() *dockertest.Resource {
	// pulls an image, creates a container based on it and runs it
	mongoContainerName := d.containerNames[wekanDBName]
	slog.Info("démarre le container wekan-db", slog.String("name", mongoContainerName))

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
		slog.Error(
			"erreur pendant le démarrage du container",
			slog.String("container", mongoContainerName),
			slog.Any("error", err),
		)
		panic(err)
	}
	// container stops after 60 seconds
	if err = wekanDB.Expire(120); err != nil {
		killContainer(wekanDB)
		slog.Error(
			"erreur pendant la mise en expiration du containe",
			slog.String("container", mongoContainerName),
			slog.Any("error", err),
		)
		panic(err)
	}
	return wekanDB
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
func wait4PostgresIsReady(datapiDBUrl string) {
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool := currentPool()
	pool.MaxWait = 120 * time.Second
	if err := pool.Retry(func() error {
		pgConnexion, err := pgx.Connect(context.Background(), datapiDBUrl)
		if err != nil {
			return err
		}
		return pgConnexion.Ping(context.Background())
	}); err != nil {
		slog.Error(
			"erreur lors de la connexion au conteneur datapi",
			slog.Any("error", err),
			slog.String("url", datapiDBUrl),
		)
		panic(err)
	}
	slog.Info("la base de données est prête", slog.String("url", datapiDBUrl))
}

func getContainer(name containerName) (*dockertest.Resource, bool) {
	return d.pool.ContainerByName(d.containerNames[name])
}

func currentPool() *dockertest.Pool {
	return d.pool
}

func GenerateRandomAPIPort() string {
	port := fake.IntBetween(30000, 33000)
	slog.Debug("nouveau port d'api", slog.Int("port", port))
	return strconv.Itoa(port)
}
