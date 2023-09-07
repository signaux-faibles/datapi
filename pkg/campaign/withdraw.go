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

func withdrawPending(ctx context.Context,
	campaignID CampaignID,
	campaignEtablissementID CampaignEtablissementID,
	username libwekan.Username,
	zone Zone,
	detail string,
) (Message, error) {
	var campaignEtablissementActionID *int
	var codeDepartement *string
	dbConn := db.Get()
	row := dbConn.QueryRow(ctx, sqlWithdrawPendingEtablissement, campaignID,
		campaignEtablissementID, username, zone, detail)
	err := row.Scan(&campaignEtablissementActionID, &codeDepartement)
	if campaignEtablissementActionID == nil || codeDepartement == nil {
		return Message{}, TakeNotFoundError{err: errors.New("aucun id retourné par l'insert")}
	} else if err != nil {
		return Message{}, err
	}
	message := Message{
		campaignID,
		campaignEtablissementID,
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

	zone := zoneForUser(libwekan.Username(s.Username))

	message, err := withdrawPending(ctx,
		CampaignID(campaignID),
		CampaignEtablissementID(campaignEtablissementID),
		libwekan.Username(s.Username),
		zone,
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
