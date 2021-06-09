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
	"io"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
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
	username := c.GetString("username")
	fceVisitURL, err := makeFceVisitURL(siret, fce.Credential, username)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	c.String(200, fceVisitURL)
}

func encrypt(stringToEncrypt string, key string) (string, error) {
	bytesToEncrypt := []byte(stringToEncrypt)
	keyBytes := []byte(key)
	cipherBlock, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	encryptedBytes := aesGCM.Seal(nonce, nonce, bytesToEncrypt, nil)
	return fmt.Sprintf("%x", encryptedBytes), nil
}

func decrypt(encryptedString string, key string) (string, error) {
	encryptedBytes, _ := hex.DecodeString(encryptedString)
	keyBytes := []byte(key)
	cipherBlock, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return "", err
	}
	nonceSize := aesGCM.NonceSize()
	nonce, data := encryptedBytes[:nonceSize], encryptedBytes[nonceSize:]
	decryptedBytes, err := aesGCM.Open(nil, nonce, data, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", decryptedBytes), nil
}

func makeFceVisitURL(siret string, credential string, username string) (string, error) {
	fceVisitURL := viper.GetString("fceVisitURL") + "establishment/" + siret
	params := url.Values{}
	params.Add("anchor", "direccte")
	params.Add("credential", credential)
	apiKey := viper.GetString("apiFceKey")
	encryptedUsername, err := encrypt(username, apiKey)
	params.Add("username", encryptedUsername)
	if err != nil {
		return "", err
	}
	fceVisitURL +=  "?" + params.Encode()
	return fceVisitURL, nil
}