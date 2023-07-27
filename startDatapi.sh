#! /usr/bin/env bash

# abort on nonzero exitstatus
set -o errexit
# abort on unbound variable
set -o nounset
# don't hide errors within pipes
set -o pipefail


function main() {

  ## Démarrage base de données
  log "démarre la base de données si nécessaire"
  POSTGRES_ID=$(docker ps --filter "name=datapiDB" --quiet)
  SHUTDOWN_POSTGRES_ID=$(docker ps --all --filter "name=datapiDB" --quiet)
  if [[ -z "${SHUTDOWN_POSTGRES_ID}" ]]; then
    log "crée le container de DB"
    docker run  --detach --tty \
                  --name      datapiDB \
                  --env       POSTGRES_USER=postgres \
                  --env       POSTGRES_PASSWORD=test \
                  --env       POSTGRES_DB=datapi_test \
                  --env       listen_addresses='*' \
                  --publish   5432:5432 \
                  --mount     "type=bind,source=$(pwd)/test/postgresql.conf,destination=/etc/postgresql/postgresql.conf" \
                  --mount     "type=bind,source=$(pwd)/test/initDB,destination=/docker-entrypoint-initdb.d" \
                  postgres:15-alpine
    log "attend que la DB soit prete (5s)"
    sleep 5
  elif [[ -z "${POSTGRES_ID}" ]]; then
    log "démarre le container de DB"
    docker start datapiDB
  else
    log "la base de données est déjà démarrée"
  fi

  ## Construction du binaire
  log "construit le binaire datapi avec les options de debug"
  go build -gcflags="-N -l" -o datapi
  log "pour debugger avec goland => 'Attach to process...'"

  ## Redémarrage de datapi
  log "redémarrage de datapi (configuration via le fichier config.toml)"
  (ps | grep datapi | grep -v grep | awk '{print $1}' | xargs -r kill -9) || log "pas besoin de stopper datapi avant le démarrage"
  ./datapi
}

function log() {
    printf '%s\n' "start datapi - $(date +%F_%T.%N) : $*"
}

function demarreDatapiDB() {
    docker run  --detach --tty \
                --name      datapiDB \
                --env       POSTGRES_USER=postgres \
                --env       POSTGRES_PASSWORD=test \
                --env       POSTGRES_DB=datapi_test \
                --env       listen_addresses='*' \
                --publish   5432:5432 \
                --mount     "type=bind,source=$(pwd)/test/postgresql.conf,destination=/etc/postgresql/postgresql.conf" \
                --mount     "type=bind,source=$(pwd)/test/initDB,destination=/docker-entrypoint-initdb.d" \
                postgres:15-alpine 2> /dev/null
    log "attends que DatapiDB soit prete"
    sleep 5
}


main "${@}"
