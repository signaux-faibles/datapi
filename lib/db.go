package dalib

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lib/pq"
	hstore "github.com/lib/pq/hstore"
	"github.com/spf13/viper"
)

func initDB() {
	var err error
	connStr := viper.GetString("postgres")
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("database connexion:" + err.Error())
	}

	_, err = db.Query(`create extension if not exists hstore`)
	if err != nil {
		log.Fatal("hstore installation: " + err.Error())
	}

	_, err = db.Query(`create sequence if not exists bucket_id`)
	if err != nil {
		log.Fatal("sequence creation: " + err.Error())
	}

	_, err = db.Query(`create table if not exists bucket ( 
		id bigint primary key default nextval('bucket_id'),
		key hstore,
		scope text[],
		value jsonb)`)
	if err != nil {
		log.Fatal("schema configuration: " + err.Error())
	}
}

// Query get objets that fits the key and the scope
func Query(key Map, scope Tags) ([]Object, error) {
	var objects []Object

	rows, err := db.Query(`
	with pairs as (
		select b.key, v.key as vkey, last(v.value) as vvalue
		from bucket b,
		lateral jsonb_each(value) v
		where b.key = $1
		and $2 @> scope
		group by b.key, v.key)
	select key, jsonb_object_agg(vkey, vvalue) from pairs
	group by key
	`, key.Hstore(), pq.Array(scope))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		o := Object{}
		var hkey hstore.Hstore

		scope := make([]sql.NullString, 0)

		err = rows.Scan(&hkey, &o.Value)

		if err != nil {
			fmt.Println("wtf")
			fmt.Println(err)
			break
		} else {
			o.Scope.FromNullStringArray(scope)
			o.Key.fromHstore(hkey)
			objects = append(objects, o)
		}
	}
	return objects, err
}

// Insert manage to insert objects in the specified bucket
func Insert(objects []Object, bucket string) error {
	tx, _ := db.Begin()

	stmt, err := tx.Prepare(`insert into bucket (key, scope, value) values ($1, $2, $3)`)
	if err != nil {
		return err
	}

	for _, o := range objects {
		o.Key["bucket"] = bucket
		_, err = stmt.Exec(o.Key.Hstore(), pq.Array(o.Scope), o.Value)
	}

	tx.Commit()

	return err
}
