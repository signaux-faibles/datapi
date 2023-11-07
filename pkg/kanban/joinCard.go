package kanban

import (
	"context"
	"github.com/signaux-faibles/libwekan"
)

func (service wekanService) JoinCard(ctx context.Context, cardID libwekan.CardID, userID libwekan.UserID) error {
	_, err := wekan.EnsureMemberInCard(ctx, cardID, userID)
	return err
}
