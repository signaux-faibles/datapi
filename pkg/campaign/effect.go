package campaign

import (
	"context"
	"datapi/pkg/core"
	"github.com/signaux-faibles/libwekan"
)

type CampaignEffect interface {
	Do(ctx context.Context) error
}

func buildEffect(effect actionEffect, user libwekan.User, kanbanService core.KanbanService, campaignEtablissementID CampaignEtablissementID) CampaignEffect {
	switch effect.Type {
	case "follow":

		return CampaignFollowEffect{
			cardID:                  effect.CardID,
			user:                    user,
			kanbanService:           kanbanService,
			campaignEtablissementID: campaignEtablissementID,
		}

	case "nofollow":

		return CampaignNofollowEffect{
			cardID:        effect.CardID,
			user:          user,
			kanbanService: kanbanService,
		}

	default:
		return CampaignNilEffect{}
	}
}

type CampaignNilEffect struct{}

func (c CampaignNilEffect) Do(ctx context.Context) error {
	return nil
}

type CampaignFollowEffect struct {
	cardID                  libwekan.CardID
	user                    libwekan.User
	kanbanService           core.KanbanService
	campaignEtablissementID CampaignEtablissementID
}

func (c CampaignFollowEffect) Do(ctx context.Context) error {
	if c.cardID == "" {
		card, _, err := upsertCard(ctx, c.campaignEtablissementID, "inscrire ici les informations de cet accompagnement", c.kanbanService, c.user.Username)
		if err != nil {
			return err
		}
		c.cardID = card.ID
	}
	card, err := c.kanbanService.SelectCardFromCardID(ctx, c.cardID, c.user.Username)
	if err != nil {
		return err
	}
	err = c.kanbanService.JoinCard(ctx, card.ID, c.user)
	if err != nil {
		return err
	}
	return nil
}

type CampaignNofollowEffect struct {
	cardID        libwekan.CardID
	user          libwekan.User
	kanbanService core.KanbanService
}

func (c CampaignNofollowEffect) Do(ctx context.Context) error {
	card, err := c.kanbanService.SelectCardFromCardID(ctx, c.cardID, c.user.Username)
	if err != nil {
		return err
	}
	config := c.kanbanService.LoadConfigForUser(c.user.Username)
	var listID libwekan.ListID
	for id, list := range config.Boards[card.BoardID].Lists {
		if list.Title == "Pas d'accompagnement" {
			listID = id
		}
	}

	if config.Boards[card.BoardID].Lists[card.ListID].Title == "Analyse en cours" && listID != "" {
		return c.kanbanService.MoveCardList(ctx, card.ID, listID, c.user)
	}

	return libwekan.ListNotFoundError{}
}
