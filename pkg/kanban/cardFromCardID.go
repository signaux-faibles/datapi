package kanban

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/utils"
	"github.com/signaux-faibles/libwekan"
)

// SelectCardFromCardID returns a KanbanCard object only if user is allowed to go on board
func (service wekanService) SelectCardFromCardID(ctx context.Context, cardID libwekan.CardID, username libwekan.Username) (core.KanbanCard, error) {
	config := service.LoadConfigForUser(username)
	wekanCard, err := wekan.GetCardWithCommentsFromID(ctx, cardID)

	boardIDs := utils.GetKeys(config.Boards)
	if utils.Contains(boardIDs, wekanCard.Card.BoardID) {
		card := wekanCardWithCommentsToKanbanCard(username)(wekanCard)
		return card, err
	}

	return core.KanbanCard{}, core.ForbiddenError{}
}

func (service wekanService) GetCardMembers(c context.Context, id libwekan.CardID, username string) ([]libwekan.Activity, error) {
	return nil, nil
}
