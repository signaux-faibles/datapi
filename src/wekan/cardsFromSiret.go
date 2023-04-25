package wekan

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/datapi/src/utils"
	"github.com/signaux-faibles/libwekan"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
)

func selectCardsFromSiret(ctx context.Context, siret string, username libwekan.Username) ([]core.KanbanCard, error) {
	pipeline := wekan.BuildCardFromCustomTextFieldPipeline("SIRET", siret)

	lookupCommentsStage := bson.M{
		"$lookup": bson.M{
			"from":         "card_comments",
			"localField":   "_id",
			"foreignField": "cardId",
			"as":           "lastActivity",
		},
	}

	replaceDateLastActivityStage := bson.M{
		"$addFields": bson.M{
			"dateLastActivity": bson.M{
				"$max": bson.A{
					bson.M{"$max": "$lastActivity.modifiedAt"},
					"$dateLastActivity",
				},
			},
		},
	}

	removeLastActivityStage := bson.M{
		"$project": bson.M{
			"lastActivity": false,
		},
	}

	pipeline = append(pipeline, lookupCommentsStage, replaceDateLastActivityStage, removeLastActivityStage)

	cards, err := wekan.SelectCardsFromPipeline(ctx, "customFields", pipeline)
	if err != nil {
		return nil, err
	}
	a, _ := json.MarshalIndent(pipeline, "", "  ")
	fmt.Println(string(a))
	return utils.Convert(cards, wekanCardToKanbanCard(username)), nil
}

func wekanCardURL(wekanCard libwekan.Card) string {
	return viper.GetString("wekanURL") + "b/" + string(wekanCard.BoardID) + "/" + string(wekanConfig.Boards[wekanCard.BoardID].Board.Slug)
}

func wekanCardToKanbanCard(username libwekan.Username) func(libwekan.Card) core.KanbanCard {
	return func(wekanCard libwekan.Card) core.KanbanCard {
		user, _ := wekanConfig.GetUserByUsername(username)
		boardConfig := wekanConfig.Boards[wekanCard.BoardID]

		card := core.KanbanCard{
			ListTitle:    boardConfig.Lists[wekanCard.ListID].Title,
			Archived:     false,
			BoardTitle:   boardConfig.Board.Title,
			Creator:      wekanConfig.Users[wekanCard.UserID].Username,
			LastActivity: wekanCard.DateLastActivity,
			StartAt:      wekanCard.StartAt,
			EndAt:        wekanCard.EndAt,
			ID:           wekanCard.ID,
		}

		if boardConfig.Board.UserIsActiveMember(user) {
			card.ListID = wekanCard.ListID
			card.BoardID = wekanCard.BoardID
			card.CreatorID = wekanCard.UserID
			card.AssigneeIDs = wekanCard.Assignees
			card.LabelIDs = wekanCard.LabelIDs
			card.MemberIDs = wekanCard.Members
			card.Description = wekanCard.Description
			card.URL = wekanCardURL(wekanCard)
			card.UserIsBoardMember = true
		}

		return card
	}
}
