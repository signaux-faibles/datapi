#!/bin/bash
echo "Attention ce script ne doit pas être utilisé en production"
if [ "$1" != "OuiJeVeuxToutCasser" ]
then
  echo "usage: mount_testdata.sh OuiJeVeuxToutCasser"
  exit
fi

./reset.sh OuiJeVeuxToutCasser
zcat ../test/initDB/testData.sql.gz|docker exec -i postgres /usr/local/bin/psql -U postgres datapiDev
