package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"net/http"
	"strconv"
)

func (c CampaignEtablissementActionID) Tuple() []interface{} {
	return []interface{}{&c}
}

func doAction(ctx context.Context,
	ids IDs,
	username string,
	action Action,
) (Message, error) {
	var campaignEtablissementActionID *int
	var codeDepartement *string
	zone := zoneForUser(username)
	dbConn := db.Get()
	row := dbConn.QueryRow(ctx, sqlActionMyEtablissement, ids.CampaignID,
		ids.CampaignEtablissementID, username, zone, action.action, action.detail)
	err := row.Scan(&campaignEtablissementActionID, &codeDepartement)
	if campaignEtablissementActionID == nil || codeDepartement == nil {
		return Message{}, TakeNotFoundError{err: errors.New("aucun id retourné par l'insert")}
	} else if err != nil {
		return Message{}, err
	}
	message := Message{
		ids.CampaignID,
		&ids.CampaignEtablissementID,
		[]string{*codeDepartement},
		action.action,
		string(username),
	}
	return message, nil
}

func actionHandlerFunc(actionLabel string, kanbanService core.KanbanService) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		var s core.Session
		s.Bind(ctx)

		var params actionParam
		err := ctx.Bind(&params)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, "décodage de la requete impossible: "+err.Error())
		}
		if params.Detail == "" {
			ctx.JSON(http.StatusBadRequest, "le paramètre detail doit contenir au moins 1 caractère")
		}

		campaignID, err := strconv.Atoi(ctx.Param("campaignID"))
		if err != nil {
			ctx.JSON(400, `/campaign/`+actionLabel+`/:campaignID/:campaignEtablissementID: le parametre campaignID doit être un entier`)
			return
		}

		campaignEtablissementID, err := strconv.Atoi(ctx.Param("campaignEtablissementID"))
		if err != nil {
			ctx.JSON(400, `/campaign/`+actionLabel+`/:campaignID/:campaignEtablissementID: le parametre campaignEtablissementID doit être un entier`)
			return
		}

		ids := IDs{
			CampaignID(campaignID),
			CampaignEtablissementID(campaignEtablissementID),
		}

		action := Action{
			action: actionLabel,
			detail: params.Detail,
		}

		message, err := doAction(ctx,
			ids,
			s.Username,
			action,
		)

		if errors.As(err, &TakeNotFoundError{}) {
			ctx.JSON(http.StatusUnprocessableEntity, "traitement indisponible pour cet établissement")
			return
		} else if err != nil {
			ctx.JSON(http.StatusInternalServerError, "erreur innattendue")
			return
		}

		user, ok := kanbanService.GetUser(libwekan.Username(s.Username))
		if !ok {
			ctx.JSON(400, "nom d'utilisateur non présent dans la base")
		}

		if params.Effect != nil {
			effect := buildEffect(*params.Effect, user, kanbanService)
			err = effect.Do(ctx)
			if err != nil {
				stream.Message <- message
				ctx.JSON(http.StatusOK, "partial ok")
				return
			}
		}

		stream.Message <- message
		ctx.JSON(http.StatusOK, "ok")
	}
}
