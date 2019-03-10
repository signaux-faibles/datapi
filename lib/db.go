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
// when tx is nil, a new transaction is started and commited.
func Query(bucket string, params QueryParams, userScope Tags, applyPolicies bool, tx *sql.Tx) ([]Object, error) {
	if tx == nil {
		var err error
		tx, err = db.Begin()
		if err != nil {
			return nil, err
		}
		defer tx.Commit()
	}
	var objects []Object

	if _, ok := params.Key["bucket"]; ok {
		return nil, errors.New("'bucket' subkey is reserved by datapi")
	}

	if bucket == "" {
		return nil, errors.New("illegal empty bucket name")
	}
	var read, promote Tags
	if applyPolicies {
		read, _, promote = ApplyPolicies(CurrentBucketPolicies.safeRead(), bucket, params.Key, userScope)
	}
	fmt.Println(applyPolicies)
	fmt.Println(read, promote)
	fmt.Println(userScope)

	userScope = userScope.Union(promote)

	if userScope == nil {
		userScope = make([]string, 0)
	}

	if params.Key == nil {
		params.Key = make(Map)
	}
	params.Key["bucket"] = bucket

	rows, err := tx.Query(`
	with pairs as (
		select b.key, v.key as vkey, last(v.value order by id) as vvalue, max(add_date) as add_date
		from bucket b,
		lateral jsonb_each(value) v
		where b.key @> $1
		and $2 @> (coalesce(scope, '{}') || $4)
		and release_date <= least($3, current_timestamp) 
		group by b.key, v.key)
	select key, jsonb_object_agg(vkey, vvalue), max(pairs.add_date) add_date 
	from pairs
	where vvalue is not null
	group by key
	`, params.Key.Hstore(), pq.Array(userScope), params.Date, pq.Array(read))
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
			return nil, errors.New("critical: unreadable data from database")
		}
		o.Scope.FromNullStringArray(scope)
		o.Key.fromHstore(hkey)
		objects = append(objects, o)

	}
	return objects, err
}

// ErrInvalidScope is returned when a query looks for an unauthorized scope
var ErrInvalidScope = errors.New("invalid scope")

// ErrInvalidPolicy is throwed when an object aiming to be a policy can't be build
var ErrInvalidPolicy = errors.New("invalid policy object")

// ErrInvalidUser is throwed when an object aiming to be an user doesn't conform
var ErrInvalidUser = errors.New("invalid user object")

// Insert manage to insert objects in the specified bucket
func Insert(objects []Object, bucket string, userScope Tags, author string) error {
	tx, _ := db.Begin()
	var refreshPolicy = false
	policyKey := Map{
		"type":   "policy",
		"bucket": "system",
	}

	stmt, err := tx.Prepare(`insert into bucket (key, scope, value, author) values ($1, $2, $3, $4)`)
	if err != nil {
		return err
	}

	for _, o := range objects {
		_, write, promote := ApplyPolicies(CurrentBucketPolicies.safeRead(), bucket, o.Key, userScope)
		fmt.Println(write, promote)
		if userScope.Union(promote).Contains(write) {
			o.Key["bucket"] = bucket

			if o.Key.contains(policyKey) {
				_, err := BuildPolicy(o)
				if err != nil {
					return ErrInvalidPolicy
				}
				refreshPolicy = true
			}
			_, err = stmt.Exec(o.Key.Hstore(), pq.Array(o.Scope), o.Value, author)
			if err != nil {
				tx.Rollback()
				return err
			}
		} else {
			tx.Rollback()
			return ErrInvalidScope
		}
	}

	if refreshPolicy {
		cbp, err := LoadPolicies(nil, tx)
		if err != nil {
			tx.Rollback()
			return err
		}
		CurrentBucketPolicies.safeUpdate(cbp)
	}

	tx.Commit()

	return nil
}
