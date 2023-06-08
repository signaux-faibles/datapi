package core

import (
	"datapi/pkg/db"
	"datapi/pkg/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"net/http"
)

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

	if !utils.Contains(s.roles, "wekan") {
		getEtablissementsFollowedByCurrentUser(c)
		return
	}

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

	params.BoardIDs = kanban.ClearBoardIDs(params.BoardIDs, params.User)

	cards, err := kanban.SelectFollowsForUser(c, params, db.Get(), s.roles)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	if len(cards.Summaries) > 0 {
		c.JSON(200, cards)
	} else {
		c.JSON(204, []string{})
	}
}

func kanbanNewCardHandler(c *gin.Context) {
	var s session
	s.Bind(c)
	var params KanbanNewCardParams
	err := c.Bind(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
	}
	if params.Labels == nil {
		params.Labels = make([]libwekan.BoardLabelName, 0)
	}
	err = kanban.CreateCard(c, params, libwekan.Username(s.Username), db.Get())
	if errors.As(err, &ForbiddenError{}) {
		c.JSON(http.StatusForbidden, err.Error())
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
	}
}
