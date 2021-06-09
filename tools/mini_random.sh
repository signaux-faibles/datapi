#!/noshell
# aide mémoire
docker exec -it postgres /usr/local/bin/dropdb -U postgres datapiDev; docker exec -it postgres /usr/local/bin/createdb -U postgres datapiDev
killall datapi
cd ..
./datapi &
pid=$!
sleep 10
http :3000/utils/import
kill $pid
docker exec -i postgres /usr/local/bin/psql -U postgres datapiDev < ./tools/randomize_database.sql
docker exec -i postgres /usr/local/bin/psql -U postgres datapiDev < ./tools/small_database.sql
docker exec -it postgres /usr/local/bin/pg_dump -U postgres datapiDev |gzip > ./test/data/testData.sql.gz
cd tools
