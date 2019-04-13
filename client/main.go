package daclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	dalib "github.com/signaux-faibles/datapi/lib"
)

// DatapiServer connects and sends objects to a datapi server
type DatapiServer struct {
	URL      string `json:"-"`
	Email    string `json:"email"`
	Password string `json:"password"`
	token    string
	Timeout  time.Duration
}

// IsConnected checks token validity
func (ds *DatapiServer) IsConnected() bool {
	if ds.token != "" {
		return true
	}
	return false
}

// QueryParams is a proxy type to dalib.QueryParams to avoid loading dalib when using client
type QueryParams struct {
	Key    map[string]string `json:"key"`
	Limit  *int              `json:"limit"`
	Offset *int              `json:"offset"`
	Date   *int              `json:"date"`
}

// PropertyMap are datas
type PropertyMap map[string]interface{}

// Object is a proxy type to dalib.Object to avoid loading dalib when using client
type Object struct {
	ID          int               `json:"id,omitempty"`
	Key         map[string]string `json:"key"`
	Scope       []string          `json:"scope,omitempty"`
	AddDate     *time.Time        `json:"addDate,omitempty"`
	ReleaseDate *time.Time        `json:"releaseDate,omitempty"`
	Value       PropertyMap       `json:"value"`
}

// Put sends objects to Datapi Server
func (ds *DatapiServer) Put(bucket string, objects []Object) error {
	return nil
}

// Get retrieves objects from Datapi Server according to Query
func (ds *DatapiServer) Get(bucket string, query QueryParams) ([]dalib.Object, error) {
	queryBody, err := json.Marshal(query)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("POST", ds.URL+"data/get/"+bucket, bytes.NewBuffer(queryBody))
	req.Header.Add("Authorization", "Bearer "+ds.token)
	data, err := client.Do(req)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(data.Body)

	var response []dalib.Object
	fmt.Println(string(buf.Bytes()))
	err = json.Unmarshal(buf.Bytes(), &response)

	return response, err
}

// Connect returns a datapi client with a token
func (ds *DatapiServer) Connect(user string, password string) error {
	jsonCredentials, err := json.Marshal(map[string]string{
		"email":    user,
		"password": password,
	})

	if err != nil {
		return err
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	loginRequest, err := client.Post(
		ds.URL+"login",
		"application/json",
		bytes.NewBuffer((jsonCredentials)),
	)

	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(loginRequest.Body)
	var result map[string]string
	err = json.Unmarshal(body, &result)

	if _, ok := result["token"]; err != nil && !ok {
		return err
	}

	ds.token = result["token"]

	return err
}
