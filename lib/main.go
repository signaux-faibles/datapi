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

// Tags is a string slice with a validation method.
type Tags []*string

// Hstore returns an hstore with data of Map
func (m Map) Hstore() hstore.Hstore {
	return MapToHstore(m)
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

func init() {
	initDB()
}

// Value transform propertyMap type to a database driver compatible type
func (p PropertyMap) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

func createDB() {

}
