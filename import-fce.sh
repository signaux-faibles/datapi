#!/bin/bash

# Met à jour la colonne `visite_fce` de la table `etablissement`, en fonction
# de la liste de SIRETs d'établissements ayant fait l'objet d'une visite,
# fournie par l'API de FCE.
#
# Test: import-fce-test.sh

set -e # stop on any failure

trap "rm fce-download.csv" EXIT

if [ ! -f fce-download.csv ]; then
  echo ""
  echo "🌍 Téléchargement des visites ..."
  curl \
    -H "Authorization: Apikey ${FCE_API_KEY}" \
    -G https://dgefp.opendatasoft.com/api/records/1.0/download/ \
    --data-urlencode "dataset=fce-visites" >fce-download.csv
fi

echo ""
echo "🚚 Importation temporaire des SIRETs de fce-download.csv dans la base de données ..."
cat fce-download.csv | PGPASSWORD="${DB_PASSWORD}" psql -U "${DB_USER}" "${DB_NAME:-datapi}" -c "CREATE TABLE tmp_fce_sirets (siret VARCHAR(14)); COPY tmp_fce_sirets FROM stdin WITH (FORMAT csv);"

echo ""
echo "📬 Mise à jour du champ visite_fce des etablissements ..."
PGPASSWORD="${DB_PASSWORD}" psql -U "${DB_USER}" "${DB_NAME:-datapi}" << COMMANDS
  UPDATE etablissement SET visite_fce = EXISTS(
    SELECT siret FROM tmp_fce_sirets WHERE siret=etablissement.siret
  );
  DROP TABLE tmp_fce_sirets;
  SELECT visite_fce, count(*) FROM etablissement GROUP BY visite_fce;
COMMANDS
