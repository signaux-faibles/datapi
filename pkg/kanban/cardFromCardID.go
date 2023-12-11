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

func wekanToKanbanJoinActivities(wekanActivities []libwekan.Activity) []core.KanbanActivity {
	var kanbanActivities []core.KanbanActivity
	var mapActivity = make(map[libwekan.UserID]core.KanbanActivity)
	for _, wekanActivity := range wekanActivities {
		if wekanActivity.ActivityType == "joinMember" || wekanActivity.ActivityType == "joinAssignee" {
			from := wekanActivity.ModifiedAt
			mapActivity[wekanActivity.MemberID] = core.KanbanActivity{
				MemberID: wekanActivity.MemberID,
				From:     &from,
			}
		} else if wekanActivity.ActivityType == "unjoinMember" || wekanActivity.ActivityType == "unjoinAssignee" {
			activity := mapActivity[wekanActivity.MemberID]
			to := wekanActivity.ModifiedAt
			activity.To = &to
			kanbanActivities = append(kanbanActivities, activity)
			delete(mapActivity, wekanActivity.MemberID)
		}
	}
	for _, activity := range mapActivity {
		kanbanActivities = append(kanbanActivities, activity)
	}
	return kanbanActivities
}

func (service wekanService) GetCardMembersHistory(c context.Context, cardID libwekan.CardID, username string) ([]core.KanbanActivity, error) {
	wekanActivities, err := wekan.SelectActivitiesFromCardID(c, cardID)
	return wekanToKanbanJoinActivities(wekanActivities), err
}
