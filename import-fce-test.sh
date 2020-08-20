#!/bin/bash

# Test de import-fce.sh
#
# Usage:
# 1. $ docker run --name sf-postgres -P -p 127.0.0.1:5432:5432 -e POSTGRES_PASSWORD=mysecretpassword -d postgres:10
# 2. $ docker cp import-fce.sh sf-postgres:/
# 3. $ docker cp import-fce-test.sh sf-postgres:/
# 4. $ docker exec -it sf-postgres /import-fce-test.sh

set -e # stop on any failure

trap "[ -f fce-download.csv ] && rm fce-download.csv" EXIT

if [ ! -f fce-download.csv.bak ]; then
  cp fce-download.csv fce-download.csv.bak || true
fi

echo ""
echo "👻 Création de données de tests ..."

cat >fce-download.csv << CONTENT
siret
12345678901234
12345678901236
CONTENT

psql -U "${DB_USER:-postgres}" << COMMANDS
DROP DATABASE IF EXISTS datapi_test;
CREATE DATABASE datapi_test;
\connect datapi_test
DROP DATABASE IF EXISTS etablissement;
CREATE TABLE etablissement (siret VARCHAR(14), visite_fce BOOLEAN);
INSERT INTO etablissement (siret) VALUES ('12345678901234');
INSERT INTO etablissement (siret, visite_fce) VALUES ('12345678901235', true);
INSERT INTO etablissement (siret, visite_fce) VALUES ('12345678901236', false);
COMMANDS

RESULTS=$(psql -U "${DB_USER:-postgres}" "datapi_test" -c "SELECT visite_fce, count(*) FROM etablissement GROUP BY visite_fce;")
echo "${RESULTS}" | grep --quiet " f          |     1" # test will fail if count does not match
echo "${RESULTS}" | grep --quiet " t          |     1" # test will fail if count does not match
echo "${RESULTS}" | grep --quiet "            |     1" # test will fail if count does not match
echo "${RESULTS}" | grep --quiet "(3 rows)"            # test will fail if count does not match

DB_USER="${DB_USER:-postgres}" DB_NAME="datapi_test" ./import-fce.sh

RESULTS=$(psql -U "${DB_USER:-postgres}" "datapi_test" -c "SELECT visite_fce, count(*) FROM etablissement GROUP BY visite_fce;")
echo "${RESULTS}" | grep --quiet " f          |     1" # test will fail if count does not match
echo "${RESULTS}" | grep --quiet " t          |     2" # test will fail if count does not match
echo "${RESULTS}" | grep --quiet "(2 rows)"            # test will fail if count does not match

echo "✅ import-fce.sh fonctionne comme prévu"
