#!/bin/bash
echo "Attention ce script ne doit pas être utilisé en production"
if [ "$1" != "OuiJeVeuxToutCasser" ] 
then
  echo "usage: reset.sh OuiJeVeuxToutCasser"
  exit
fi

echo "Arrêt pgadmin"
docker stop pgadmin > /dev/null 2>&1 
echo "Suppression de la base datapiDev"
docker exec -it postgres /usr/local/bin/dropdb -U postgres datapiDev > /dev/null 2>&1
echo "Création de la base datapiDev à vide"
docker exec -it postgres /usr/local/bin/createdb -U postgres datapiDev > /dev/null 2>&1
echo "Démarrage pgadmin"
docker start pgadmin > /dev/null 2>&1 
