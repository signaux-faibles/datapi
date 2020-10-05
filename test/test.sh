#!/bin/sh

TEST_DATA = "data/testData.sql.gz"
POSTGRES_PASS = "test"
POSTGRES_DOCKERPORT = "65432"

# instancie postgres et charge la base de données
POSTGRES_CONTAINER=$(docker run -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD -p $PG_DOCKERPORT -d postgres:11.9)







# arrete le docker postgres et supprime le container