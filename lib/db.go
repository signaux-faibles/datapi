package dalib

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lib/pq"
	hstore "github.com/lib/pq/hstore"
)

func initDB() {
	var err error
	connStr := "user=postgres dbname=test"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("database connexion:" + err.Error())
	}

	_, err = db.Query(`create extension if not exists hstore`)
	if err != nil {
		log.Fatal("hstore installation: " + err.Error())
	}
	_, err = db.Query(`create table if not exists bucket ( 
		id bigint primary key default nextval('bucket_id'),
		key hstore,
		scope text[],
		value jsonb)`)
	if err != nil {
		log.Fatal("schema configuration: " + err.Error())
	}
	_, err = db.Query(`create sequence if not exists bucket_id`)
	if err != nil {
		log.Fatal("sequence configuration: " + err.Error())
	}
}

// Query get objets that fits the key and the scope
func Query(key Map, scope Tags) ([]Object, error) {
	var objects []Object

	rows, err := db.Query(`
		SELECT id, key, scope, value
		FROM bucket
		WHERE key @> $1
		and scope <@ $2
		ORDER BY id`, key.Hstore(), pq.Array(scope))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		o := Object{}
		var hkey hstore.Hstore
		rows.Next()
		scope := make([]sql.NullString, 0)

		err = rows.Scan(&o.ID, &hkey, pq.Array(&scope), &o.Value)

		if err != nil {
			fmt.Println(err)
			break
		} else {
			o.Scope.FromNullStringArray(scope)
			o.Key.fromHstore(hkey)
			objects = append(objects, o)
		}
	}
	return objects, nil
}

// Insert manage to insert objects in the specified bucket
func Insert(objects []Object) error {
	stmt, err := db.Prepare(`insert into bucket (key, scope, value) values ($1, $2, $3)`)
	if err != nil {
		return err
	}

	for _, o := range objects {
		_, err = stmt.Exec(o.Key.Hstore(), pq.Array(o.Scope), o.Value)
	}

	return err
}
