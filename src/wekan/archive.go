package wekan

import (
	"context"
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/libwekan"
)

func (s wekanService) UnarchiveCard(ctx context.Context, cardID libwekan.CardID, username libwekan.Username) error {
	card, err := wekan.GetCardFromID(ctx, cardID)
	if err != nil {
		return core.UnknownCardError{"cardID=" + string(cardID)}
	}
	user, ok := wekanConfig.GetUserByUsername(username)
	if !ok {
		return core.ForbiddenError{"l'utilisateur n'est pas habilité à désarchiver cette carte"}
	}
	board, ok := wekanConfig.Boards[card.BoardID]
	if !ok {
		return core.UnknownBoardError{"boardID=" + string(card.BoardID)}
	}
	if !board.Board.UserIsActiveMember(user) {
		return core.ForbiddenError{"l'utilisateur n'est pas habilité à désarchiver cette carte"}
	}
	return wekan.UnarchiveCard(ctx, cardID)
}
