package dalib

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq" // postgresql driver
	hstore "github.com/lib/pq/hstore"
	"github.com/spf13/viper"
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
	ID          int         `json:"id,omitempty"`
	Key         Map         `json:"key"`
	Scope       Tags        `json:"scope,omitempty"`
	AddDate     *time.Time  `json:"addDate,omitempty"`
	ReleaseDate *time.Time  `json:"releaseDate,omitempty"`
	Value       PropertyMap `json:"value"`
}

// Map is equivalent to hstore in a nice json fashion.
type Map map[string]string

// Set writes a value to a Map
func (m *Map) Set(key string, value string) {
	if *m == nil {
		*m = make(map[string]string)
	}
	(*m)[key] = value
}

// Get reads a value from a Map
func (m *Map) Get(key string) (string, error) {
	if *m == nil {
		return "", errors.New("key " + key + " not set")
	}
	if _, ok := (*m)[key]; !ok {
		return "", errors.New("key " + key + " not set")
	}
	return (*m)[key], nil
}

// Tags is a string slice with a validation method.
type Tags []string

// Contains tests if a Tag collection is contained in another Tag collection
func (t Tags) Contains(t2 Tags) bool {
	mt := make(map[string]struct{})
	for _, tag := range t {
		mt[tag] = struct{}{}
	}
	for _, tag := range t2 {
		if _, ok := mt[tag]; !ok {
			return false
		}
	}
	return true
}

// fromHstore writes Hstore values to Map
func (m *Map) fromHstore(h hstore.Hstore) {
	if *m == nil {
		*m = make(map[string]string)
	}
	for k, v := range h.Map {
		if v.Valid {
			s := v.String
			(*m)[k] = s
		}
	}
}

// FromNullStringArray converts []*string to []sql.NullString
// replaces original Tags object
func (t *Tags) FromNullStringArray(array []sql.NullString) {
	*t = make(Tags, 0)
	for _, v := range array {
		if v.Valid {
			s := v.String
			*t = append(*t, s)
		}
	}
}

// Hstore returns an hstore with data of Map
func (m Map) Hstore() hstore.Hstore {
	var h hstore.Hstore
	h.Map = make(map[string]sql.NullString)
	for k, v := range m {
		h.Map[k] = sql.NullString{
			String: v,
			Valid:  true,
		}
	}
	return h
}

// ToScope dirty way to convert data from DB to Tags #TODO: clean that
func ToScope(claim []interface{}) Tags {
	t := make([]string, 0)
	for _, s := range claim {
		t = append(t, s.(string))
	}
	return t
}

// ToMap dirty way to convert data from DB to Maps #TODO: clean that
func ToMap(value interface{}) (Map, error) {
	var m Map
	j, err := json.Marshal(value)
	if err != nil {
		return m, err
	}
	err = json.Unmarshal(j, &m)
	if err != nil {
		return m, err
	}

	// if the map is empty, we want {}, not null
	if m == nil {
		m = make(Map)
	}
	return m, nil
}

// Bucket contains informations needed to process queries and apply some permissions
type Bucket struct {
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
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	fmt.Println(viper.GetString("postgres"))

	if err != nil {
		panic(err)
	}

	initDB()
	CurrentBucketPolicies = LoadPolicies(time.Now())
}

// Value transform propertyMap type to a database driver compatible type
func (p PropertyMap) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

func createDB() {

}
