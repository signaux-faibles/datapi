package dalib

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	hstore "github.com/lib/pq/hstore"
)

// CreateUser creates a datapi user (when there's no one)
func CreateUser(connStr, user, password string) error {
	Warmup(connStr)
	cursor := db.QueryRow("select case when count(*)=0 then true else false end from bucket")
	var empty bool
	cursor.Scan(&empty)

	if empty {
		log.Print("database is empty: creating user is allowed")
		user := Object{
			Key: Map{
				"email": user,
				"type":  "credentials",
			},
			Value: PropertyMap{
				"password": password,
				"scope":    Tags{"system-reader", "system-writer"},
			},
		}
		systemPolicy := Object{
			Key: Map{
				"type": "policy",
				"name": "system",
			},
			Value: PropertyMap{
				"match": "system",
				"key":   Map{},
				"scope": Tags{},
				"read":  Tags{"system-reader"},
				"write": Tags{"system-writer"},
			},
		}

		err := Insert([]Object{user, systemPolicy}, "system", nil, "system")
		if err != nil {
			log.Fatal("Error creating user: " + err.Error())
		}
	} else {
		log.Fatal("Database is not empty, creating such user is not allowed")
	}
	log.Printf("user '%s' created with password '%s'", user, password)
	return nil
}

// CreateDB creates a brand new (and empty) database
func CreateDB(dbname, connStr string) {
	var err error

	pgConnStr := connStr + " dbname='postgres'"
	db, err = sql.Open("postgres", pgConnStr)
	if err != nil {
		log.Fatal("database connexion:" + err.Error())
	}
	log.Print("Connected to postgres database")

	_, err = db.Query(fmt.Sprintf(`create database %s`, dbname))
	if err != nil {
		log.Fatal("database creation: " + err.Error())
	}
	log.Printf("%s database created", dbname)

	db.Close()

	dbConnStr := connStr + fmt.Sprintf(" dbname='%s'", dbname)
	db, err = sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatal("database switch: " + err.Error())
	}

	log.Printf("switched to new database: %s", dbname)

	_, err = db.Query(`create extension if not exists hstore`)
	if err != nil {
		log.Fatal("hstore installation: " + err.Error())
	}
	log.Print("hstore extension installed")

	_, err = db.Query(`create extension if not exists pgcrypto;`)
	if err != nil {
		log.Fatal("pgcrypto installation: " + err.Error())
	}
	log.Print("pgcrypto extension installed")

	_, err = db.Query(`CREATE OR REPLACE FUNCTION public.last_agg(
    anyelement,
    anyelement)
  RETURNS anyelement AS
	$BODY$
    SELECT $2;
	$BODY$
  LANGUAGE sql IMMUTABLE STRICT
	COST 100;`)
	if err != nil {
		log.Fatal("last_agg creation: " + err.Error())
	}
	log.Print("last_agg function created")

	_, err = db.Query(`CREATE AGGREGATE last(anyelement) (
		SFUNC=last_agg,
		STYPE=anyelement
	)`)
	if err != nil {
		log.Fatal("last creation: " + err.Error())
	}
	log.Print("last aggregate created")

	_, err = db.Query(`DROP AGGREGATE IF EXISTS array_cat_agg(anyarray)`)
	if err != nil {
		log.Fatal("array_cat_agg creation: " + err.Error())
	}
	_, err = db.Query(`CREATE AGGREGATE array_cat_agg(anyarray) (
		SFUNC=array_cat,
		STYPE=anyarray
	)`)
	if err != nil {
		log.Fatal("array_cat_agg creation: " + err.Error())
	}
	log.Print("array_cat_agg aggregate created")

	_, err = db.Query(`create sequence if not exists bucket_id`)
	if err != nil {
		log.Fatal("bucket sequence creation: " + err.Error())
	}
	log.Print("bucket_id sequence created")

	_, err = db.Query(`create table if not exists bucket ( 
		id bigint primary key default nextval('bucket_id'),
		key hstore,
		author text,
		add_date timestamp default current_timestamp,
		release_date timestamp default current_timestamp,
		scope text[],
		value jsonb)`)
	if err != nil {
		log.Fatal("schema configuration: " + err.Error())
	}
	log.Print("bucket table created")

	_, err = db.Query(`create sequence if not exists logs_id`)
	if err != nil {
		log.Fatal("logs_id sequence creation: " + err.Error())
	}
	log.Print("logs sequence created")

	_, err = db.Query(`create table if not exists logs (
		id bigint primary key default nextval('logs_id'),
		reader text,
		read_date timestamp,
		query jsonb
	)`)
	if err != nil {
		log.Fatal("logs table creation: " + err.Error())
	}
	log.Print("logs table sequence created")
	log.Print("database successfully created")
	db.Close()
}

func initDB(connStr string) {
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("database connexion:" + err.Error())
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

	if _, ok := params.Key["bucket"]; ok || bucket == "" {
		return nil, ErrInvalidBucket
	}

	var bp BucketPolicies
	if applyPolicies {
		bp = RelevantPolicies(CurrentBucketPolicies.SafeRead(), bucket, userScope)
	}

	if bp == nil {
		bp = make(BucketPolicies, 0)
	}

	jsonbp, err := json.Marshal(bp)
	if err != nil {
		return nil, errors.New("shit in the fan")
	}
	if userScope == nil {
		userScope = make([]string, 0)
	}

	if params.Key == nil {
		params.Key = make(Map)
	}
	params.Key["bucket"] = bucket

	query := `
	with policy_data as (
		select b.value->'name' as name,
					 b.value->'key' as key, 
					 b.value->'read' as read, 
					 b.value->'write' as write, 
					 b.value->'promote' as promote 
		from jsonb_array_elements($4::jsonb) b
		),
	policy_jsonb as (
		SELECT p.name, hstore(array_agg(b.key), array_agg(b.value)) as key, p.read, p.write, p.promote
		FROM policy_data p,
		lateral jsonb_each_text(p.key) b
		group by p.name, p.read, p.write, p.promote),
	policy as (select p.name, p.key, read.read, write.write, promote.promote  from policy_jsonb p
		cross join lateral (
			select array_agg(d.value) AS read
			from jsonb_array_elements_text(p.read) d
		) read
		cross join lateral (
			select array_agg(d.value) AS write
			from jsonb_array_elements_text(p.write) d
		) write
		cross join lateral (
			select array_agg(d.value) AS promote
			from jsonb_array_elements_text(p.promote) d
		) promote),
	bucket_with_policy as (
		select b.*, array_cat_agg(p.read) as policy_read, array_cat_agg(p.write) as policy_write, array_cat_agg(p.promote) as policy_promote from bucket b
		left join policy p on p.key <@ b.key
		where b.key @> $1
		group by b.id),
	pairs as (
		select b.key, v.key as vkey, last(v.value order by id) as vvalue, max(add_date) as add_date
		from bucket_with_policy b,
		lateral jsonb_each(value) v
		where $2 || coalesce(b.policy_promote, '{}') @> (coalesce(b.scope, '{}') || (coalesce(b.policy_read, '{}')) )
		and release_date <= least($3, current_timestamp) 
		group by b.key, v.key)
	select key, jsonb_object_agg(vkey, vvalue), max(pairs.add_date) add_date 
	from pairs 
	where vvalue is not null
	group by key
	`

	if params.Limit != nil && *params.Limit > 0 {
		query = query + fmt.Sprintf(" limit %d", *params.Limit)
	}

	if params.Offset != nil {
		query = query + fmt.Sprintf(" offset %d", *params.Offset)
	}

	rows, err := tx.Query(query, params.Key.Hstore(), pq.Array(userScope), params.Date, jsonbp)
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
		if o.Value == nil {
			return ErrInvalidObject
		}

		_, write, promote, writable := ApplyPolicies(CurrentBucketPolicies.SafeRead(), bucket, o.Key, userScope)
		for k := range o.Value {
			if writable != nil && !writable.Contains(Tags{k}) {
				return ErrForbiddenValue
			}
		}
		if userScope.Union(promote).Contains(write) {
			if o.Key == nil {
				o.Key = make(Map)
			}
			o.Key["bucket"] = bucket

			if o.Key.Contains(policyKey) {
				_, err := BuildPolicy(o)
				if err != nil {
					return ErrInvalidPolicy
				}
				refreshPolicy = true
			}

			for _, t := range triggers {
				if o.Key.Contains(t.Key) {
					o, err = t.Function(o)
					if err != nil {
						return err
					}
				}
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
		CurrentBucketPolicies.SafeUpdate(cbp)
	}

	tx.Commit()

	return nil
}
