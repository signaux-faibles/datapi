package core

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/signaux-faibles/datapi/src/utils"
	"net/http"
	"regexp"
)

func getCodesNaf(c *gin.Context) {
	rows, err := db.Get().Query(context.Background(), "select code, libelle from naf where niveau=1")
	if err != nil {
		utils.AbortWithError(c, err)
		return
	}
	var naf = make(map[string]string)
	for rows.Next() {
		var code string
		var libelle string
		err := rows.Scan(&code, &libelle)
		if err != nil {
			utils.AbortWithError(c, err)
			return
		}
		naf[code] = libelle
	}
	c.JSON(http.StatusOK, naf)
}

func departementsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, departements)
}

func regionsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, regions)
}

func validSiren(c *gin.Context) {
	siren := c.Param("siren")
	match, err := regexp.MatchString("^[0-9]{9}$", siren)
	if err != nil || !match {
		c.AbortWithStatusJSON(400, "SIREN valide obligatoire")
	}
	c.Next()
}

func validSiret(c *gin.Context) {
	siren := c.Param("siret")
	match, err := regexp.MatchString("^[0-9]{14}$", siren)
	if err != nil || !match {
		c.AbortWithStatusJSON(http.StatusBadRequest, "SIRET valide obligatoire")
	}
	c.Next()
}

type session struct {
	username string
	auteur   string
	roles    Scope
}

func (s *session) bind(c *gin.Context) {
	s.username = c.GetString("username")
	s.auteur = c.GetString("given_name") + " " + c.GetString("family_name")
	s.roles = scopeFromContext(c)
}

func (s session) hasRole(role string) bool {
	return utils.Contains(s.roles, role)
}
