package kanban

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/utils"
	"github.com/signaux-faibles/libwekan"
)

func (service wekanService) SelectCardFromCardID(ctx context.Context, cardID libwekan.CardID, username libwekan.Username) (core.KanbanCard, error) {
	config := service.LoadConfigForUser(username)
	wekanCard, err := wekan.GetCardFromID(ctx, cardID)

	boardIDs := utils.GetKeys(config.Boards)
	if utils.Contains(boardIDs, wekanCard.BoardID) {
		card := wekanCardToKanbanCard(username)(wekanCard)
		return card, err
	}

	return core.KanbanCard{}, core.ForbiddenError{}
}
