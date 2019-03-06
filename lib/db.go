package dalib

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

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

	_, err = db.Query(`create extension if not exists pgcrypto;`)
	if err != nil {
		log.Fatal("pgcrypto installation: " + err.Error())
	}

	_, err = db.Query(`create sequence if not exists bucket_id`)
	if err != nil {
		log.Fatal("sequence creation: " + err.Error())
	}

	_, err = db.Query(`create table if not exists bucket ( 
		id bigint primary key default nextval('bucket_id'),
		key hstore,
		add_date timestamp default current_timestamp,
		release_date timestamp default current_timestamp,
		scope text[],
		value jsonb)`)
	if err != nil {
		log.Fatal("schema configuration: " + err.Error())
	}
}

// QueryParams contains informations needed to perform a read query
type QueryParams struct {
	Key    Map        `json:"key"`
	Limit  *int       `json:"limit"`
	Offset *int       `json:"offset"`
	Date   *time.Time `json:"date"`
}

// Query get objets that fits the key and the scope
func Query(bucket string, params QueryParams, userScope Tags) ([]Object, error) {
	var objects []Object

	if _, ok := params.Key["bucket"]; ok {
		return nil, errors.New("illegal bucket in key")
	}

	if bucket == "" {
		return nil, errors.New("empty bucket name is not allowed")
	}

	bucketReadScope := ApplyReadPolicies(CurrentBucketPolicies, bucket, params.Key)
	if params.Key == nil {
		params.Key = make(Map)
	}
	params.Key["bucket"] = bucket

	rows, err := db.Query(`
	with pairs as (
		select b.key, v.key as vkey, last(v.value) as vvalue, max(add_date) as add_date
		from bucket b,
		lateral jsonb_each(value) v
		where b.key @> $1
		and $2 @> (coalesce(scope, '{}') || $4)
		and release_date <= greatest($3, current_timestamp) 
		group by b.key, v.key)
	select key, jsonb_object_agg(vkey, vvalue), max(pairs.add_date) add_date 
	from pairs
	where vvalue is not null
	group by key
	`, params.Key.Hstore(), pq.Array(userScope), params.Date, pq.Array(bucketReadScope))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		o := Object{}
		var hkey hstore.Hstore

		scope := make([]sql.NullString, 0)

		err = rows.Scan(&hkey, &o.Value, &o.AddDate)

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

// ErrInvalidScope is returned when a query looks for an unauthorized scope
var ErrInvalidScope = errors.New("invalid scope")

// Insert manage to insert objects in the specified bucket
func Insert(objects []Object, bucket string, user User) error {
	tx, _ := db.Begin()

	stmt, err := tx.Prepare(`insert into bucket (key, scope, value) values ($1, $2, $3)`)
	if err != nil {
		return err
	}

	for _, o := range objects {
		scope := ApplyWritePolicies(CurrentBucketPolicies, bucket, o.Key)

		if user.Scope.Contains(scope) {
			o.Key["bucket"] = bucket
			_, err = stmt.Exec(o.Key.Hstore(), pq.Array(o.Scope), o.Value)
		} else {
			tx.Rollback()
			return ErrInvalidScope
		}
	}

	tx.Commit()

	return err
}
