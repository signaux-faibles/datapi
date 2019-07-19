package dalib

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/lib/pq"
	hstore "github.com/lib/pq/hstore"
)

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
	Key    Map            `json:"key"`
	Limit  *int           `json:"limit"`
	Offset *int           `json:"offset"`
	Filter []FilterParams `json:"filter"`
	Sort   []struct {
		Field string `json:"field"`
		Order int    `json:"order"`
	} `json:"sort"`
	Date *time.Time `json:"date"`
}

// FilterParams handles line filtering parameters
type FilterParams struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// Prepare caches a query in a faster temporary table (poor optimization, but efficient)
func Prepare(bucket string, userScope Tags, tx *sql.Tx, owner string) error {
	testOwner, err := regexp.Compile("[,(); ]")
	if err != nil {
		return err
	}
	if testOwner.MatchString(owner) {
		return errors.New("unallowed characters in owner string, aborting")
	}

	if tx == nil {
		var err error
		tx, err = db.Begin()
		if err != nil {
			return err
		}
		defer tx.Commit()
	}

	if bucket == "" {
		return ErrInvalidBucket
	}

	bp := RelevantPolicies(CurrentBucketPolicies.SafeRead(), bucket, userScope)
	jsonbp, err := json.Marshal(bp)
	if err != nil {
		return errors.New("bucket policy error, something bad is happening")
	}
	if userScope == nil {
		userScope = make([]string, 0)
	}

	query := `
	create table if not exists "cache_` + owner + `" as
	with policy_data as (
		select b.value->'name' as name,
					 b.value->'key' as key, 
					 b.value->'read' as read, 
					 b.value->'write' as write, 
					 b.value->'promote' as promote 
		from jsonb_array_elements($4::jsonb) b
		),
	policy_jsonb as (
		SELECT p.name, hstore(
			array_remove(array_agg(b.key),null), 
			array_remove(array_agg(b.value),null)
		) as key, p.read, p.write, p.promote
		FROM policy_data p
		left join lateral jsonb_each_text(p.key) b on true
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
		group by b.id
	),
	pairs as (
		select b.key, v.key as vkey, last(v.value order by id) as vvalue, max(add_date) as add_date
		from bucket_with_policy b,
		lateral jsonb_each(value) v
		where $2 || coalesce(b.policy_promote, '{}') @> (coalesce(b.scope, '{}') || (coalesce(b.policy_read, '{}')) )
		and release_date <= least($3, current_timestamp) 
		group by b.key, v.key),
	values as (
		select key, jsonb_object_agg(vkey, vvalue) as values, max(pairs.add_date) add_date 
	  from pairs
	  where vvalue is not null
		group by key)
	select key, values, add_date from values;`

	query2 := `create index if not exists "x_` + owner + `"  ON "cache_` + owner + `" (
		(values->>'score')
	);`

	query3 := `create index if not exists "x_` + owner + `" ON "cache_` + owner + `" 
	using gist(key);
  `

	key := Map{
		"bucket": bucket,
		"type":   "detection",
	}
	_, err = tx.Exec(query, key.Hstore(), pq.Array(userScope), time.Now(), jsonbp)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec(query2)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec(query3)
	if err != nil {
		tx.Rollback()
		return err
	}

	return err
}

// Query get objets that fits the key and the scope
// when tx is nil, a new transaction is started and commited.
func Query(bucket string, params QueryParams, userScope Tags, applyPolicies bool, tx *sql.Tx, reader string) ([]Object, error) {
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

	var bp = make(BucketPolicies, 0)
	if applyPolicies {
		bp = RelevantPolicies(CurrentBucketPolicies.SafeRead(), bucket, userScope)
	}

	jsonbp, err := json.Marshal(bp)
	if err != nil {
		return nil, errors.New("bucket policy error, something bad is happening")
	}
	if userScope == nil {
		userScope = make([]string, 0)
	}

	if params.Key == nil {
		params.Key = make(Map)
	}
	params.Key["bucket"] = bucket

	// logging query
	_, err = tx.Exec("insert into logs (reader, read_date, query) values ($1, current_timestamp, $2);", reader, params)
	if err != nil {
		return nil, errors.New("Couldn't log the query, contact support")
	}
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
		SELECT p.name, hstore(
			array_remove(array_agg(b.key),null), 
			array_remove(array_agg(b.value),null)
		) as key, p.read, p.write, p.promote
		FROM policy_data p
		left join lateral jsonb_each_text(p.key) b on true
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
		group by b.id
	),
	pairs as (
		select b.key, v.key as vkey, last(v.value order by id) as vvalue, max(add_date) as add_date
		from bucket_with_policy b,
		lateral jsonb_each(value) v
		where $2 || coalesce(b.policy_promote, '{}') @> (coalesce(b.scope, '{}') || (coalesce(b.policy_read, '{}')) )
		and release_date <= least($3, current_timestamp) 
		group by b.key, v.key),
	values as (
		select key, jsonb_object_agg(vkey, vvalue) as values, max(pairs.add_date) add_date 
	  from pairs
	  where vvalue is not null
		group by key)
	select key, values, add_date from values`

	if len(params.Sort) == 1 {
		if params.Sort[0].Field == "score" {
			query = query + "	order by (values->>'score')::float8"
			if params.Sort[0].Order == -1 {
				query = query + "	desc"
			}
		}
	}
	rows, err := tx.Query(query, params.Key.Hstore(), pq.Array(userScope), params.Date, jsonbp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offset, limit int
	if params.Offset != nil {
		offset = *params.Offset
	}
	if params.Limit != nil {
		limit = *params.Limit
	}
	for rows.Next() {
		if offset > 0 {
			offset--
			continue
		}

		if params.Limit == nil || limit > 0 {

			o := Object{}
			var hkey hstore.Hstore

			scope := make([]sql.NullString, 0)
			err = rows.Scan(&hkey, &o.Value, &o.AddDate)
			if err != nil {
				return nil, errors.New("critical: unreadable data from database")
			}

			o.Scope.FromNullStringArray(scope)
			o.Key.fromHstore(hkey)
			if filter(o, params.Filter) {
				if filterGreen(o) {
					limit--
				}
				objects = append(objects, o)
			}
		}
	}

	return objects, err
}

func filter(o Object, params []FilterParams) bool {
	for _, p := range params {
		f, ok := filters[p.Field]
		if !ok {
			return false
		}
		r := f(o, p.Operator, p.Value)
		if !r {
			return false
		}
	}
	return true
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

// Cache get objets that fits the key and the scope from the prepared cache
// when tx is nil, a new transaction is started and commited.
func Cache(bucket string, params QueryParams, userScope Tags, applyPolicies bool, tx *sql.Tx, token string, reader string) ([]Object, error) {
	testOwner, err := regexp.Compile("[,(); ]")
	if err != nil {
		return nil, err
	}
	if testOwner.MatchString(reader) {
		return nil, errors.New("unallowed characters in owner string, aborting")
	}
	if tx == nil {
		var err error
		tx, err = db.Begin()
		if err != nil {
			return nil, err
		}
		defer tx.Commit()
	}

	err = Prepare(bucket, userScope, tx, reader)
	if err != nil {
		return nil, err
	}
	var objects []Object

	if _, ok := params.Key["bucket"]; ok || bucket == "" {
		return nil, ErrInvalidBucket
	}

	if params.Key == nil {
		params.Key = make(Map)
	}
	params.Key["bucket"] = bucket
	// logging query
	_, err = tx.Exec("insert into logs (reader, read_date, query) values ($1, current_timestamp, $2);", token, params)
	if err != nil {
		return nil, errors.New("Couldn't log the query, contact support")
	}
	query := `select * from "cache_` + reader + `"
	where key @> $1`

	if len(params.Sort) == 1 {
		if params.Sort[0].Field == "score" {
			query = query + "	order by values->>'score'"
			if params.Sort[0].Order == -1 {
				query = query + "	desc"
			}
		}
	}
	rows, err := tx.Query(query, params.Key.Hstore())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offset, limit int
	if params.Offset != nil {
		offset = *params.Offset
	}
	if params.Limit != nil {
		limit = *params.Limit
	}
	for rows.Next() {
		if offset > 0 {
			offset--
			continue
		}

		if params.Limit == nil || limit > 0 {

			o := Object{}
			var hkey hstore.Hstore

			scope := make([]sql.NullString, 0)
			err = rows.Scan(&hkey, &o.Value, &o.AddDate)
			if err != nil {
				return nil, errors.New("critical: unreadable data from database")
			}

			o.Scope.FromNullStringArray(scope)
			o.Key.fromHstore(hkey)
			if filter(o, params.Filter) {
				if filterGreen(o) {
					limit--
				}
				objects = append(objects, o)
			}
		}
	}

	return objects, err
}
