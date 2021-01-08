package main

import (
	"encoding/json"
	"github.com/spf13/viper"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Fce type pour l'API Fiche Commune Entreprise
type Fce struct {
	Success		bool 	`json:"success"`
	Credential 	string 	`json:"credential"`
}

func askFceCredentials() ([]byte, error) {
	apiFceURL := viper.GetString("apiFceURL")
	req, err := http.NewRequest("GET", apiFceURL + "askCredential", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	now := time.Now() 
	ts := strconv.FormatInt(now.Unix(), 10)
    q.Add("ts", ts)
	apiKey := viper.GetString("apiFceKey")
	q.Add("api_key", apiKey)
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func getFceURL(c *gin.Context) {
	siret := c.Param("siret")
	if siret == "" {
		c.JSON(400, "SIRET obligatoire")
		return
	}
	body, err := askFceCredentials()
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	fce := Fce{}
	bodyString := string(body)
	err = json.Unmarshal([]byte(bodyString), &fce)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	fceVisitURL := makeFceVisitURL(siret, fce.Credential)
	c.String(200, fceVisitURL)
}

func makeFceVisitURL(siret string, credential string) (string) {
	fceVisitURL := viper.GetString("fceVisitURL") + "establishment/" + siret
	params := url.Values{}
	params.Add("credential", credential)
	fceVisitURL +=  "?" + params.Encode()
	return fceVisitURL
}