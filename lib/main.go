package dalib

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"regexp"
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

// fromHstore converts hstore to Map
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

// ToMap secure way to cast untrusted object to Map
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

	if err != nil {
		panic(err)
	}

	initDB()
	cbp, err := LoadPolicies(nil, nil)
	if err != nil {
		panic(err)
	}
	CurrentBucketPolicies.SafeUpdate(cbp)
}

// Value transform propertyMap type to a database driver compatible type
func (p PropertyMap) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

// ToScope secure way to cast untrusted object to Tags
func ToScope(scope interface{}) (Tags, error) {
	j, err := json.Marshal(scope)
	if err != nil {
		return nil, err
	}

	var s Tags
	err = json.Unmarshal(j, &s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Union of two Tags produces a new Tags containing all tags.
func (t Tags) Union(t2 Tags) (newTags Tags) {
	m := make(map[string]struct{})

	for _, v := range t {
		m[v] = struct{}{}
	}
	for _, v := range t2 {
		m[v] = struct{}{}
	}
	for k := range m {
		newTags = append(newTags, k)
	}
	return newTags
}

// DAError carries error message and http code
type DAError struct {
	Message string
	Code    int
}

func (e DAError) Error() string {
	return e.Message
}

// ErrInvalidScope is returned when a query looks for an unauthorized scope
var ErrInvalidScope = DAError{"invalid scope", 403}

// ErrForbiddenValue is throwed when an object contains keys an user is'nt allowed to put
var ErrForbiddenValue = DAError{"forbidden value", 403}

// ErrInvalidPolicy is throwed when an object aiming to be a policy can't be build
var ErrInvalidPolicy = DAError{"invalid policy object", 400}

// ErrInvalidBucket is throwed when the specified bucket is illegal
var ErrInvalidBucket = DAError{"invalid bucket name", 400}

// ErrInvalidUser is throwed when an object aiming to be an user doesn't conform
var ErrInvalidUser = DAError{"invalid user object", 400}

// ErrInvalidObject is throwed when an object can't be processed by main insert process or trigger
var ErrInvalidObject = DAError{"invalid data object", 400}

// ErrInvalidEmail is throwed when a user object is invalid
var ErrInvalidEmail = DAError{"invalid email adress", 400}

var rxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
