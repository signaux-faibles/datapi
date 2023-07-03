package core

import (
	"context"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"datapi/pkg/db"
	"datapi/pkg/utils"
)

func getCodesNaf(c *gin.Context) {
	rows, err := db.Get().Query(context.Background(), "select code, libelle from naf where niveau=1")
	defer rows.Close()
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
	c.JSON(http.StatusOK, Departements)
}

func regionsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, Regions)
}

func checkSirenFormat(c *gin.Context) {
	siren := c.Param("siren")
	match, err := regexp.MatchString("^[0-9]{9}$", siren)
	if err != nil || !match {
		c.AbortWithStatusJSON(http.StatusBadRequest, "SIREN valide obligatoire")
	}
	c.Next()
}

func checkSiretFormat(c *gin.Context) {
	siren := c.Param("siret")
	match, err := regexp.MatchString("^[0-9]{14}$", siren)
	if err != nil || !match {
		c.AbortWithStatusJSON(http.StatusBadRequest, "SIRET valide obligatoire")
	}
	c.Next()
}

// Session correspond aux informations du user de la Session
type Session struct {
	Username string
	Auteur   string
	Roles    Scope
}

// TODO transformer cette méthode en constructeur
func (s *Session) Bind(c *gin.Context) {
	s.Username = c.GetString("username")
	s.Auteur = c.GetString("given_name") + " " + c.GetString("family_name")
	s.Roles = scopeFromContext(c)
}

func (s Session) hasRole(role string) bool {
	return utils.Contains(s.Roles, role)
}
