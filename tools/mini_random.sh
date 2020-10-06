# aide mémoire
docker exec -it postgres /usr/local/bin/dropdb -U postgres datapi; docker exec -it postgres /usr/local/bin/createdb -U postgres datapi
killall datapi
cd ..
./datapi &
pid=$!
sleep 3
http :3000/utils/import
kill $pid
docker exec -i postgres /usr/local/bin/psql -U postgres datapi < ./tools/small_database.sql
docker exec -i postgres /usr/local/bin/psql -U postgres datapi < ./tools/randomize_database.sql
docker exec -it postgres /usr/local/bin/pg_dump -U postgres datapi |gzip > ./test/data/testData.sql.gz