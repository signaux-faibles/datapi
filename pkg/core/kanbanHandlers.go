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
	var s Session
	s.Bind(c)
	kanbanConfig := Kanban.LoadConfigForUser(libwekan.Username(s.Username))
	c.JSON(http.StatusOK, kanbanConfig)
}

func kanbanGetCardsHandler(c *gin.Context) {
	var s Session
	s.Bind(c)
	siret := c.Param("siret")

	cards, err := Kanban.SelectCardsFromSiret(c, siret, libwekan.Username(s.Username))
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
	var s Session
	s.Bind(c)

	if !utils.Contains(s.Roles, "wekan") {
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
	params.User, ok = Kanban.GetUser(libwekan.Username(s.Username))
	if !ok {
		c.JSON(http.StatusForbidden, "le nom d'utilisateur n'est pas reconnu")
		return
	}

	params.BoardIDs = Kanban.ClearBoardIDs(params.BoardIDs, params.User)

	cards, err := Kanban.SelectFollowsForUser(c, params, db.Get(), s.Roles)

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
	var s Session
	s.Bind(c)
	var params KanbanNewCardParams
	err := c.Bind(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
	}
	err = Kanban.CreateCard(c, params, libwekan.Username(s.Username), db.Get())
	if errors.As(err, &ForbiddenError{}) {
		c.JSON(http.StatusForbidden, err.Error())
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
	}
}

func kanbanUpdateCardHandler(c *gin.Context) {
	var s Session
	s.Bind(c)
	var params struct {
		CardID      libwekan.CardID `json:"cardID"`
		Description string          `json:"description"`
	}
	c.Bind(&params)

	card, err := Kanban.SelectCardFromCardID(c, params.CardID, libwekan.Username(s.Username))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "erreur:" + err.Error()})
	}
	err = Kanban.UpdateCard(c, card, params.Description, libwekan.Username(s.Username))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "erreur: " + err.Error()})
	}
}

func kanbanUnarchiveCardHandler(c *gin.Context) {
	var s Session
	s.Bind(c)

	cardID := libwekan.CardID(c.Param("cardID"))

	err := Kanban.UnarchiveCard(c, cardID, libwekan.Username(s.Username))
	if errors.Is(err, UnknownCardError{}) {
		c.JSON(http.StatusNotFound, err.Error())
	}
	if errors.Is(err, ForbiddenError{}) {
		c.JSON(http.StatusForbidden, err.Error())
	}
	if errors.Is(err, UnknownBoardError{}) {
		c.JSON(http.StatusNotFound, err.Error())
	}
	if errors.Is(err, libwekan.NothingDoneError{}) {
		c.JSON(http.StatusNoContent, err.Error())
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
	}
	c.JSON(200, "traitement effectué")
}
