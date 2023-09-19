package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func withdrawPending(ctx context.Context, ids IDs, username string, detail string) (Message, error) {
	var campaignEtablissementActionID *int
	var codeDepartement *string

	zone := zoneForUser(username)

	dbConn := db.Get()
	row := dbConn.QueryRow(ctx, sqlWithdrawPendingEtablissement, ids.CampaignID,
		ids.CampaignEtablissementID, username, zone, detail)
	err := row.Scan(&campaignEtablissementActionID, &codeDepartement)

	if campaignEtablissementActionID == nil || codeDepartement == nil {
		return Message{}, TakeNotFoundError{err: errors.New("aucun id retourné par l'insert")}
	} else if err != nil {
		return Message{}, err
	}

	message := Message{
		ids.CampaignID,
		ids.CampaignEtablissementID,
		[]string{*codeDepartement},
		"withdraw",
		string(username),
	}

	return message, nil
}

func withdrawHandler(ctx *gin.Context) {
	var s core.Session
	s.Bind(ctx)

	var params actionParam
	err := ctx.Bind(&params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, "décodage de la requete impossible: "+err.Error())
		return
	}
	if params.Detail == "" {
		ctx.JSON(http.StatusBadRequest, "le paramètre detail doit contenir au moins 1 caractère")
		return
	}

	campaignID, err := strconv.Atoi(ctx.Param("campaignID"))
	if err != nil {
		ctx.JSON(400, `/campaign/withdraw/:campaignID/:campaignEtablissementID: le parametre campaignID doit être un entier`)
		return
	}

	campaignEtablissementID, err := strconv.Atoi(ctx.Param("campaignEtablissementID"))
	if err != nil {
		ctx.JSON(400, `/campaign/withdraw/:campaignID/:campaignEtablissementID: le parametre campaignEtablissementID doit être un entier`)
		return
	}

	ids := IDs{
		CampaignID(campaignID),
		CampaignEtablissementID(campaignEtablissementID),
	}

	message, err := withdrawPending(ctx,
		ids,
		s.Username,
		params.Detail)

	if errors.As(err, &TakeNotFoundError{}) {
		ctx.JSON(http.StatusUnprocessableEntity, "traitement indisponible pour cet établissement")
		return
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, "erreur innattendue")
		return
	}

	stream.Message <- message
	ctx.JSON(http.StatusOK, "ok")
}
