package core

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/datapi/src/db"
	"github.com/signaux-faibles/datapi/src/utils"
	"github.com/signaux-faibles/libwekan"
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

func kanbanConfigHandler(c *gin.Context) {
	var s session
	s.Bind(c)
	kanbanConfig := kanban.LoadConfigForUser(libwekan.Username(s.Username))
	c.JSON(http.StatusOK, kanbanConfig)
}

func kanbanGetCardsHandler(c *gin.Context) {
	var s session
	s.Bind(c)
	siret := c.Param("siret")

	cards, err := kanban.SelectCardsFromSiret(c, siret, libwekan.Username(s.Username))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	if len(cards) > 0 {
		c.JSON(200, cards)
	} else {
		c.JSON(204, cards)
	}
}

func kanbanGetCardsForCurrentUserHandler(c *gin.Context) {
	var s session
	s.Bind(c)

	var params = KanbanSelectCardsForUserParams{}
	err := c.Bind(&params)
	if err != nil {
		c.JSON(400, err.Error())
		return
	}

	types := []string{"no-card", "my-cards", "all-cards"}
	if !utils.Contains(types, params.Type) {
		c.AbortWithStatusJSON(400, fmt.Sprintf("`%s` n'est pas un type supporté", params.Type))
		return
	}

	var ok bool
	params.User, ok = kanban.GetUser(libwekan.Username(s.Username))
	if !ok {
		c.JSON(http.StatusForbidden, "le nom d'utilisateur n'est pas reconnu")
		return
	}

	cards, err := kanban.SelectCardsForUser(c, params, db.Get(), s.roles)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	if len(cards) > 0 {
		c.JSON(200, cards)
	} else {
		c.JSON(204, []string{})
	}
}

// session correspond aux informations du user de la session
type session struct {
	Username string
	auteur   string
	roles    Scope
}

func (s *session) Bind(c *gin.Context) {
	s.Username = c.GetString("username")
	s.auteur = c.GetString("given_name") + " " + c.GetString("family_name")
	s.roles = scopeFromContext(c)
}

func (s session) hasRole(role string) bool {
	return utils.Contains(s.roles, role)
}
