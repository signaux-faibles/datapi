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

type AccessLog struct {
	Path   string
	Method string
	Body   []byte
	Token  string
}

func (infos AccessLog) String() string {
	return fmt.Sprintf("log -> %s - %s - %s - %s", infos.Path, infos.Method, infos.Body, "[***]")
}

// AccessLogSaver handler pour définir une façon de sauver les logs
type AccessLogSaver func(log AccessLog) error

// LogMiddleware définit le middleware qui gère les logs
func (datapi *Datapi) LogMiddleware(c *gin.Context) {
	if c.Request.Body == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "request has nil body"})
		return
	}
	accessLog, err := extractAccessLogFrom(c)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	err = datapi.saveAccessLog(accessLog)
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
}
func extractAccessLogFrom(c *gin.Context) (AccessLog, error) {
	path := c.Request.URL.Path
	method := c.Request.Method
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return AccessLog{}, errors.Wrap(err, "erreur pendant la lecture du body de la requête")
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var token []string
	if viper.GetBool("enableKeycloak") {
		token, err = getRawToken(c)
		if err != nil {
			return AccessLog{}, errors.Wrap(err, "erreur pendant la récupération du token")
		}
	} else {
		token = []string{"", "fakeKeycloak"}
	}
	return AccessLog{
		Path:   path,
		Method: method,
		Body:   body,
		Token:  token[1],
	}, nil
}

func PrintLogToStdout(message AccessLog) error {
	log.Printf("%s", message)
	return nil
}
