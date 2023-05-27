package kanban

import (
	"context"
	"datapi/pkg/core"
	"github.com/signaux-faibles/libwekan"
)

func (service wekanService) UnarchiveCard(ctx context.Context, cardID libwekan.CardID, username libwekan.Username) error {
	card, err := wekan.GetCardFromID(ctx, cardID)
	if err != nil {
		return core.UnknownCardError{CardIdentifier: "cardID=" + string(cardID)}
	}
	user, ok := wekanConfig.GetUserByUsername(username)
	if !ok {
		return core.ForbiddenError{Reason: "l'utilisateur n'est pas habilité à désarchiver cette carte"}
	}
	board, ok := wekanConfig.Boards[card.BoardID]
	if !ok {
		return core.UnknownBoardError{BoardIdentifier: "boardID=" + string(card.BoardID)}
	}
	if !board.Board.UserIsActiveMember(user) {
		return core.ForbiddenError{Reason: "l'utilisateur n'est pas habilité à désarchiver cette carte"}
	}
	return wekan.UnarchiveCard(ctx, cardID)
}
