package core

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"datapi/pkg/db"
	"datapi/pkg/utils"
)

type HTTPCall struct {
	path   string
	method string
	body   []byte
	token  []string
}

// SaveHTTPCall handler pour définir une façon de sauver les logs
type SaveHTTPCall func(log HTTPCall) error

// LogMiddleware définit le middleware qui gère les logs
func (datapi *Datapi) LogMiddleware(c *gin.Context) {
	if c.Request.Body == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "request has nil body"})
		return
	}
	apiCall, err := extractAPICallInfosFrom(c)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	err = datapi.saveAPICall(apiCall)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}

}

func extractAPICallInfosFrom(c *gin.Context) (HTTPCall, error) {
	path := c.Request.URL.Path
	method := c.Request.Method
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return HTTPCall{}, errors.Wrap(err, "erreur pendant la lecture du body de la requête")
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var token []string
	if viper.GetBool("enableKeycloak") {
		token, err = getRawToken(c)
		if err != nil {
			return HTTPCall{}, errors.Wrap(err, "erreur pendant la récupération du token")
		}
	} else {
		token = []string{"", "fakeKeycloak"}
	}
	return HTTPCall{
		path:   path,
		method: method,
		body:   body,
		token:  token,
	}, nil
}

func printToStdout(message HTTPCall) error {
	log.Printf("log -> %s - %s - %s - %s\n", message.path, message.method, message.body, message.token[1])
	return nil
}

func saveToDB(message HTTPCall) error {
	_, err := db.Get().Exec(
		context.Background(),
		`insert into logs (path, method, body, token) values ($1, $2, $3, $4);`,
		message.path,
		message.method,
		string(message.body),
		message.token[1],
	)
	return err
}
