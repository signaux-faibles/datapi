package core

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"datapi/pkg/utils"
)

type LogInfos struct {
	Path   string
	Method string
	Body   []byte
	Token  []string
}

func (infos LogInfos) String() string {
	token := "empty"
	if len(infos.Token) > 1 && len(infos.Token[1]) > 0 {
		token = "[***]"
	}
	return fmt.Sprintf("log -> %s - %s - %s - %s", infos.Path, infos.Method, infos.Body, token)
}

// SaveLogInfos handler pour définir une façon de sauver les logs
type SaveLogInfos func(log LogInfos) error

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
func extractAPICallInfosFrom(c *gin.Context) (LogInfos, error) {
	path := c.Request.URL.Path
	method := c.Request.Method
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return LogInfos{}, errors.Wrap(err, "erreur pendant la lecture du body de la requête")
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var token []string
	if viper.GetBool("enableKeycloak") {
		token, err = getRawToken(c)
		if err != nil {
			return LogInfos{}, errors.Wrap(err, "erreur pendant la récupération du token")
		}
	} else {
		token = []string{"", "fakeKeycloak"}
	}
	return LogInfos{
		Path:   path,
		Method: method,
		Body:   body,
		Token:  token,
	}, nil
}

func PrintLogToStdout(message LogInfos) error {
	log.Printf("%s", message)
	return nil
}
