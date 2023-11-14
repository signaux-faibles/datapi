package kanban

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/utils"
	"github.com/signaux-faibles/libwekan"
	"time"
)

func (service wekanService) JoinCard(ctx context.Context, cardID libwekan.CardID, user libwekan.User) error {
	_, err := wekan.EnsureMemberInCard(ctx, cardID, user.ID)
	if err != nil {
		return err
	}
	// Si la board comporte une liste `Accompagnement en cours`, on déplace la carte dans cette colonne
	card, err := wekan.GetCardFromID(ctx, cardID)
	if err != nil {
		return err
	}
	config := kanbanConfigForUser(user.Username)

	// vérification que la board contient une liste «Accompagnement en cours»
	listID, _, isCampaignBoard := utils.MapFindTest(
		config.Boards[card.BoardID].Lists,
		func(listID libwekan.ListID, list core.KanbanList) bool {
			return list.Title == "Accompagnement en cours"
		})

	if !isCampaignBoard {
		return nil
	}

	err = wekan.EnsureMoveCardList(ctx, cardID, listID, user.ID)
	if err != nil {
		return err
	}

	if card.EndAt != nil {
		wekan.SetCardEndAt(ctx, cardID, nil)
	}

	return err
}

func (service wekanService) PartCard(ctx context.Context, cardID libwekan.CardID, user libwekan.User) error {
	_, err := wekan.EnsureMemberOutOfCard(ctx, cardID, user.ID)
	if err != nil {
		return err
	}
	// Si la board comporte une liste `Accompagnement en cours`, on déplace la carte dans cette colonne
	card, err := wekan.GetCardFromID(ctx, cardID)
	if err != nil {
		return err
	}
	config := kanbanConfigForUser(user.Username)

	// vérification qu'il n'y a plus d'accompagnant sur la carte
	n := len(card.Members) + len(card.Assignees)
	if n > 0 {
		return nil
	}

	// vérification que la board contient une liste «Accompagnement terminé»
	listID, _, isCampaignBoard := utils.MapFindTest(
		config.Boards[card.BoardID].Lists,
		func(listID libwekan.ListID, list core.KanbanList) bool {
			return list.Title == "Accompagnement terminé"
		})

	if !isCampaignBoard {
		return nil
	}

	err = wekan.EnsureMoveCardList(ctx, cardID, listID, user.ID)
	if err != nil {
		return err
	}
	now := time.Now()
	err = wekan.SetCardEndAt(ctx, cardID, &now)
	return err
}
