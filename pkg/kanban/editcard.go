package kanban

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/utils"
	"github.com/signaux-faibles/libwekan"
)

func (service wekanService) UpdateCard(ctx context.Context, card core.KanbanCard,
	description string, username libwekan.Username) error {
	boards := service.SelectBoardsForUsername(username)
	boardIDs := utils.Convert(boards, func(board libwekan.ConfigBoard) libwekan.BoardID { return board.Board.ID })
	if utils.Contains(boardIDs, card.BoardID) {
		return wekan.UpdateCardDescription(ctx, card.ID, description)
	}
	return core.ForbiddenError{}
}
