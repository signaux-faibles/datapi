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

// MapToHstore converts map[string]*string to hstore
func MapToHstore(m map[string]*string) hstore.Hstore {
	var h hstore.Hstore
	h.Map = make(map[string]sql.NullString)
	for k, v := range m {
		if v != nil {
			h.Map[k] = sql.NullString{
				String: *v,
				Valid:  true,
			}
		} else {
			h.Map[k] = sql.NullString{
				String: "",
				Valid:  false,
			}
		}
	}
	return h
}

// HstoreToMap converts hstore to map[string]string
func HstoreToMap(h hstore.Hstore) map[string]*string {
	m := make(map[string]*string)
	for k, v := range h.Map {
		if v.Valid {
			s := v.String
			m[k] = &s
		} else {
			m[k] = nil
		}
	}
	return m
}

// SliceToArray converts []*string to []sql.NullString
func (t *Tags) setArray(array []sql.NullString) {
	*t = make(Tags, 0)
	for _, v := range array {
		if v.Valid {
			s := v.String
			*t = append(*t, &s)
		} else {
			*t = append(*t, nil)
		}
	}
}

// Query get objets that fits the key
func Query(key Map, scope Tags) ([]Object, error) {
	var objects []Object

	rows, err := db.Query(`
		SELECT id, key, scope, value
		FROM bucket
		WHERE key @> $1
		and scope <@ $2
		ORDER BY id`, key.Hstore(), scope)
	if err != nil {
		return nil, err
	}

	for {
		o := Object{}
		var hkey hstore.Hstore
		rows.Next()
		scope := make([]sql.NullString, 0)

		err = rows.Scan(&o.ID, &hkey, pq.Array(&scope), &o.Value)

		if err != nil {
			fmt.Println(err)
			break
		} else {
			o.Scope.setArray(scope)
			o.Key = HstoreToMap(hkey)
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
