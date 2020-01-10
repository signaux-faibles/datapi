package daclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	dalib "github.com/signaux-faibles/datapi/lib"
)

// DatapiServer connects and sends objects to a datapi server
type DatapiServer struct {
	URL      string `json:"-"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bucket   string `json:"bucket"`
	token    string
	Scope    []string
	SendSize int // 0 is unlimited
	Timeout  time.Duration
	Errors   int
}

// Tags est une liste de tags avec quelques fonctions associées
type Tags []string

// Append ajoute de nouveaux tags dans un objet Tags
func (current *Tags) Append(new []string) {
	var unique = make(map[string]struct{})
	for _, n := range append(*current, new...) {
		unique[n] = struct{}{}
	}
	*current = nil
	for n := range unique {
		*current = append(*current, n)
	}
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
	Scope       Tags                   `json:"scope"`
	AddDate     *time.Time             `json:"addDate,omitempty"`
	ReleaseDate *time.Time             `json:"releaseDate,omitempty"`
	Value       map[string]interface{} `json:"value"`
}

// SecureSend permet d'assurer (un peu) l'envoi des objets en les répétant en cas d'erreur
func (ds *DatapiServer) SecureSend(objects []Object) error {
	if len(objects) > 0 {
		i := 0
		for {
			err := ds.Connect()
			if err == nil {
				break
			}
			log.Println("erreur de connexion datapi: " + err.Error())
			if i == 5 {
				log.Println("abandon: 5 erreurs consécutives")
				return err
			}

			log.Println("tentative de reconnexion: " + strconv.Itoa(i))
			time.Sleep(5 * time.Second)

			i++
		}

		i = 0
		for {
			err := ds.Put(objects)
			if err == nil {
				break
			}

			log.Println("erreur de transmission datapi: " + err.Error())
			if i == 5 {
				log.Println("abandon: 5 erreurs consécutives")
				return err
			}

			log.Println("tentative de réémission: " + err.Error())
			time.Sleep(5 * time.Second)

			i++
		}
	}

	return nil
}

// Put sends objects to Datapi Server
func (ds *DatapiServer) Put(objects []Object) error {
	for i := range objects {
		objects[i].Scope.Append(ds.Scope)
	}

	postData, err := json.Marshal(objects)

	buf := bytes.NewBuffer(postData)
	req, err := http.NewRequest("POST", ds.URL+"data/put/"+ds.Bucket, buf)
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
func (ds *DatapiServer) Connect() error {
	jsonCredentials, err := json.Marshal(map[string]string{
		"email":    ds.Email,
		"password": ds.Password,
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

// Worker est un worker asynchrone pour l'envoi de données vers datapi
func (ds *DatapiServer) Worker() (chan Object, *sync.WaitGroup) {
	var input = make(chan Object)
	var wait = new(sync.WaitGroup)
	wait.Add(1)

	go func() {
		i := 0
		var datas = []Object{}
		for data := range input {
			datas = append(datas, data)
			i++
			if i == ds.SendSize {
				err := ds.SecureSend(datas)
				if err != nil {
					ds.Errors++
				}
				i, datas = 0, nil
			}
		}
		err := ds.SecureSend(datas)
		if err != nil {
			ds.Errors++
		}
		wait.Done()
	}()

	return input, wait
}
