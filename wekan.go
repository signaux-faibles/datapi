package main

import (
	"encoding/json"
	"github.com/spf13/viper"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// WekanLogin type pour les réponses du service de login
type WekanLogin struct {
	Id				string 	`json:"id"`
	Token 			string 	`json:"token"`
	TokenExpires 	string 	`json:"tokenExpires"`
}

// WekanCreateToken type pour les réponses du service de création de token
type WekanCreateToken struct {
	Id				string 	`json:"_id"`
	AuthToken 		string 	`json:"authToken"`
}

// Token type pour faire persister en mémoire les tokens réutilisables
type Token struct {
	Value			string
	CreationDate	time.Time
}

var tokens sync.Map

func adminLogin(username string, password string) ([]byte, error) {
	wekanApiURL := viper.GetString("wekanApiURL")
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	req, err := http.NewRequest("POST", wekanApiURL + "users/login", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func createToken(userID string, adminToken string) ([]byte, error) {
	wekanApiURL := viper.GetString("wekanApiURL")
	req, err := http.NewRequest("POST", wekanApiURL + "api/createtoken/" + userID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer " + adminToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func isValidToken(token Token) (bool) {
	duration := time.Since(token.CreationDate)
	hours := duration.Hours()
	days := hours / 24
	if days < float64(viper.GetInt("wekanTokenDays")) {
		return true
	} else {
		return false
	}
}

func wekanHandler(c *gin.Context) {
	configFile := viper.GetString("wekanConfigFile")
	wekanConfig, err := ioutil.ReadFile(configFile)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	var root map[string]interface{}
	err = json.Unmarshal(wekanConfig, &root)
	users := root["users"].(map[string]interface{})
	username := c.GetString("username")
	log.Printf("username=%s", username)
	userID, ok := users[username]
	if !ok {
		log.Printf("not a wekan user")
	} else {
		var adminToken Token
		adminUsername := viper.GetString("wekanAdminUsername")
		loadedToken, ok := tokens.Load(adminUsername)
		if loadedToken != nil {
			adminToken = loadedToken.(Token)
			ok = isValidToken(adminToken)
		}
		if !ok {
			log.Printf("no valid admin token")
			adminPassword := viper.GetString("wekanAdminPassword")
			body, err := adminLogin(adminUsername, adminPassword)
			if err != nil {
				goto sendConfig
			}
			wekanLogin := WekanLogin{}
			err = json.Unmarshal(body, &wekanLogin)
			if err != nil {
				goto sendConfig
			}
			adminToken.Value = wekanLogin.Token
			now := time.Now()
			adminToken.CreationDate = now
			tokens.Store(adminUsername, adminToken)
		} else {
			log.Printf("existing admin token")
		}
		log.Printf("adminToken=%s",adminToken)
		var userToken Token
		loadedToken, ok = tokens.Load(username)
		if loadedToken != nil {
			userToken = loadedToken.(Token)
			ok = isValidToken(userToken)
		}
		if !ok {
			log.Printf("no valid user token")
			body, err := createToken(userID.(string), adminToken.Value)
			if err != nil {
				goto sendConfig
			}
			wekanCreateToken := WekanCreateToken{}
			err = json.Unmarshal(body, &wekanCreateToken)
			if err != nil {
				goto sendConfig
			}
			userToken.Value = wekanCreateToken.AuthToken
			now := time.Now()
			userToken.CreationDate = now
			tokens.Store(username, userToken)
		} else {
			log.Printf("existing user token")
		}
		log.Printf("userToken=%s", userToken)
		root["userToken"] = userToken.Value
	}
	goto sendConfig
	sendConfig:
	data, err := json.Marshal(root)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.String(200, string(data))
}