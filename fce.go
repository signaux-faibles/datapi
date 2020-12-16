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

type Fce struct {
	Success		bool 	`json:"success"`
	Credential 	string 	`json:"credential"`
}

func askFceCredentials() ([]byte, error) {
	apiFceUrl := viper.GetString("apiFceUrl")
	req, err := http.NewRequest("GET", apiFceUrl + "askCredential", nil)
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

func getFceUrl(c *gin.Context) {
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
	fceVisitUrl := makeFceVisitUrl(siret, fce.Credential)
	c.String(200, fceVisitUrl)
}

func makeFceVisitUrl(siret string, credential string) (string) {
	fceVisitUrl := viper.GetString("fceUrl") + "establishment/" + siret
	params := url.Values{}
	params.Add("credential", credential)
	fceVisitUrl +=  "?" + params.Encode()
	return fceVisitUrl
}