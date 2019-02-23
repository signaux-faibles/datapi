package dalib

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"

	_ "github.com/lib/pq" // postgresql driver
	hstore "github.com/lib/pq/hstore"
)

type attribution string

// Reader permits user to read values
const Reader = attribution("reader")

// Writer permits user to write values
const Writer = attribution("writer")

// Delete permits user to delete values
const Delete = attribution("delete")

// Object is the smallest data thing in API
type Object struct {
	ID    int         `json:"id" bson:"_id"`
	Key   Map         `json:"key" bson:"key"`
	Scope Tags        `json:"scope" bson:"scope"`
	Value PropertyMap `json:"value" bson:"value"`
}

// Map is equivalent to hstore in a nice json fashion.
type Map map[string]*string

// Set writes a value to a Map
func (m *Map) Set(key string, value string) {
	if *m == nil {
		*m = make(map[string]*string)
	}
	(*m)[key] = &value
}

// Get reads a value from a Map
func (m *Map) Get(key string) (string, error) {
	if *m == nil {
		return "", errors.New("key " + key + " not set")
	}
	if _, ok := (*m)[key]; !ok {
		return "", errors.New("key " + key + " not set")
	}
	return *(*m)[key], nil
}

// Tags is a string slice with a validation method.
type Tags []*string

// fromHstore writes Hstore values to Map
func (m *Map) fromHstore(h hstore.Hstore) {
	if *m == nil {
		*m = make(map[string]*string)
	}
	for k, v := range h.Map {
		if v.Valid {
			s := v.String
			(*m)[k] = &s
		} else {
			(*m)[k] = nil
		}
	}
}

// FromNullStringArray converts []*string to []sql.NullString
func (t *Tags) FromNullStringArray(array []sql.NullString) {
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

// Hstore returns an hstore with data of Map
func (m Map) Hstore() hstore.Hstore {
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

// Bucket contains informations needed to process queries and apply some permissions
type Bucket struct {
	Table string
	Scope map[string]attribution
}

// PropertyMap is a json object with all fields
type PropertyMap map[string]interface{}

// Scan writes database values in go variables
func (p *PropertyMap) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}

	var i interface{}
	err := json.Unmarshal(source, &i)
	if err != nil {
		return err
	}

	*p, ok = i.(map[string]interface{})
	if !ok {
		return errors.New("type assertion .(map[string]interface{}) failed")
	}

	return nil
}

var db *sql.DB
var adminScope Tags

func init() {
	initDB()
	a := "admin"
	adminScope = []*string{&a}
}

// Value transform propertyMap type to a database driver compatible type
func (p PropertyMap) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

func createDB() {

}
