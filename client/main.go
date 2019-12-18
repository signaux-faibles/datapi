package daclient

import (
	"bytes"
	"encoding/json"
	"errors"
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

// Object is a proxy type to dalib.Object to avoid loading dalib when using client
type Object struct {
	ID          int                    `json:"id,omitempty"`
	Key         map[string]string      `json:"key"`
	Scope       []string               `json:"scope"`
	AddDate     *time.Time             `json:"addDate,omitempty"`
	ReleaseDate *time.Time             `json:"releaseDate,omitempty"`
	Value       map[string]interface{} `json:"value"`
}

// Put sends objects to Datapi Server
func (ds *DatapiServer) Put(bucket string, objects []Object) error {
	postData, err := json.Marshal(objects)

	buf := bytes.NewBuffer(postData)

	req, err := http.NewRequest("POST", ds.URL+"data/put/"+bucket, buf)
	req.Header.Add("Authorization", "Bearer "+ds.token)
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{
		Timeout: 3600 * time.Second,
	}
	res, err := client.Do(req)
	if (*res).StatusCode != 200 {
		msg, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return errors.New("Erreur HTTP datapi: " + res.Status + " " + string(msg))
	}
	return err
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

	if loginRequest.StatusCode != 200 {
		return errors.New("datapi login failed")
	}

	body, err := ioutil.ReadAll(loginRequest.Body)
	var result struct {
		Token string `json:"access_token"`
	}
	err = json.Unmarshal(body, &result)

	if result.Token == "" || err != nil {
		return err
	}

	ds.token = result.Token

	return err
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
	err = json.Unmarshal(buf.Bytes(), &response)

	return response, err
}
