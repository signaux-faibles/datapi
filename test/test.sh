#!/bin/bash
# script de test e2e de datapi.
# avant de lancer ce script, il faut compiler le binaire datapi
# pour mettre à jour les golden files, utilisez le flag -u
# pour mettre l'environnement en attente et lancer les tests en parallèle, utilisez le flag -w
# les flags ne sont pas cumulables

set -e

function cleanup()
{
set +e
if [ "$DATAPI_PID" != "" ]; then
  echo "- arret datapi"
  kill "$DATAPI_PID"
fi

rm -rf workspace

echo "- arret conteneur"
docker stop "$POSTGRES_CONTAINER" > /dev/null 2>&1
echo "- suppression conteneur"
docker rm "$POSTGRES_CONTAINER" > /dev/null 2>&1
}

trap cleanup EXIT

echo "- compilation du binaire datapi"
cd ..
go build

cd test 
TEST_DATA="data/testData.sql.gz"
POSTGRES_PASSWORD="test"
POSTGRES_DOCKERPORT="65432"
DATAPI_PORT="12345"

echo "- initialisation docker postgres"
POSTGRES_CONTAINER=$(docker run -e POSTGRES_PASSWORD="$POSTGRES_PASSWORD" -p "$POSTGRES_DOCKERPORT":5432 -d postgres:10-alpine)
echo "- conteneur créé: $POSTGRES_CONTAINER"
sleep 4
echo "- creation de la base de données"
docker exec -i "$POSTGRES_CONTAINER" /usr/local/bin/createdb -U postgres datapi_test
echo "- insert des données de test"
zcat < "$TEST_DATA" | docker exec -i "$POSTGRES_CONTAINER" /usr/local/bin/psql -U postgres datapi_test > /dev/null 2>&1

echo "- generation configuration de test"
mkdir workspace 
sed "s/changemypass/$POSTGRES_PASSWORD/" config.toml.source | sed "s/changemyport/$DATAPI_PORT/" | sed "s/changemypgport/$POSTGRES_DOCKERPORT/" > workspace/config.toml

cp ../datapi workspace
cp -r ../migrations workspace
cd workspace

./datapi &

DATAPI_PID=$!
sleep 2

if [ "$1" = '-u' ]; then rm ../data/*.json.gz; GOLDEN_UPDATE=true; else GOLDEN_UPDATE=false; fi

if [ "$1" = '-w' ]; 
  then echo "Environnement en attente, commande à exécuter pour les tests"
    echo DATAPI_PORT="$DATAPI_PORT" GOLDEN_UPDATE="$GOLDEN_UPDATE" go test -v; 
    read -p "Appuyez sur entrée pour continuer"
fi
cd ../

DATAPI_PORT="$DATAPI_PORT" GOLDEN_UPDATE="$GOLDEN_UPDATE" go test -v
