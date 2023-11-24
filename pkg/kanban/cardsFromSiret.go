package kanban

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/utils"
	"fmt"
	"github.com/signaux-faibles/libwekan"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
)

func buildSelectCardsFromSiretPipeline(siret string) libwekan.Pipeline {
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
	return pipeline
}

func selectCardsFromSiret(ctx context.Context, siret string, username libwekan.Username) ([]core.KanbanCard, error) {
	pipeline := buildSelectCardsFromSiretPipeline(siret)

	cards, err := wekan.SelectCardsFromPipeline(ctx, "customFields", pipeline)
	if err != nil {
		return nil, err
	}
	return utils.Convert(cards, wekanCardToKanbanCard(username)), nil
}

func wekanCardWithCommentsURL(wekanCard libwekan.CardWithComments) string {
	return fmt.Sprintf("%sb/%s/%s/%s",
		viper.GetString("wekanURL"),
		string(wekanCard.Card.BoardID),
		string(WekanConfig.Boards[wekanCard.Card.BoardID].Board.Slug),
		wekanCard.Card.ID,
	)
}
func wekanCardURL(wekanCard libwekan.Card) string {
	return fmt.Sprintf("%sb/%s/%s/%s",
		viper.GetString("wekanURL"),
		string(wekanCard.BoardID),
		string(WekanConfig.Boards[wekanCard.BoardID].Board.Slug),
		wekanCard.ID,
	)
}

func wekanCardWithCommentsToKanbanCard(username libwekan.Username) func(libwekan.CardWithComments) core.KanbanCard {
	return func(wekanCard libwekan.CardWithComments) core.KanbanCard {
		user, _ := WekanConfig.GetUserByUsername(username)
		boardConfig := WekanConfig.Boards[wekanCard.Card.BoardID]

		siret := cardWithCommentsToSiret(WekanConfig)(wekanCard)

		card := core.KanbanCard{
			ListTitle:    boardConfig.Lists[wekanCard.Card.ListID].Title,
			Archived:     wekanCard.Card.Archived,
			BoardTitle:   boardConfig.Board.Title,
			Creator:      WekanConfig.Users[wekanCard.Card.UserID].Username,
			LastActivity: wekanCard.Card.DateLastActivity,
			StartAt:      wekanCard.Card.StartAt,
			EndAt:        wekanCard.Card.EndAt,
			ID:           wekanCard.Card.ID,
			Siret:        core.Siret(siret),
		}

		if boardConfig.Board.UserIsActiveMember(user) {
			card.ListID = wekanCard.Card.ListID
			card.BoardID = wekanCard.Card.BoardID
			card.CreatorID = wekanCard.Card.UserID
			card.AssigneeIDs = wekanCard.Card.Assignees
			card.SwimlaneID = wekanCard.Card.SwimlaneID
			card.LabelIDs = wekanCard.Card.LabelIDs
			card.MemberIDs = wekanCard.Card.Members
			card.Description = wekanCard.Card.Description
			card.URL = wekanCardWithCommentsURL(wekanCard)
			card.UserIsBoardMember = true
			card.Comments = utils.Convert(wekanCard.Comments, wekanCommentToKanbanComment)
		}

		return card
	}
}

func wekanCommentToKanbanComment(comment libwekan.Comment) core.KanbanComment {
	var kanbanComment core.KanbanComment
	kanbanComment.ID = string(comment.ID)
	kanbanComment.Comment = comment.Text
	kanbanComment.AuthorID = string(comment.UserID)
	kanbanComment.CreatedAt = comment.CreatedAt
	kanbanComment.ModifiedAt = comment.ModifiedAt
	return kanbanComment
}

func wekanCardToKanbanCard(username libwekan.Username) func(libwekan.Card) core.KanbanCard {
	return func(wekanCard libwekan.Card) core.KanbanCard {
		user, _ := WekanConfig.GetUserByUsername(username)
		boardConfig := WekanConfig.Boards[wekanCard.BoardID]

		siret, _ := WekanConfig.GetCardCustomFieldByName(wekanCard, "SIRET")

		card := core.KanbanCard{
			ListTitle:    boardConfig.Lists[wekanCard.ListID].Title,
			Archived:     wekanCard.Archived,
			BoardTitle:   boardConfig.Board.Title,
			Creator:      WekanConfig.Users[wekanCard.UserID].Username,
			LastActivity: wekanCard.DateLastActivity,
			StartAt:      wekanCard.StartAt,
			EndAt:        wekanCard.EndAt,
			ID:           wekanCard.ID,
			Siret:        core.Siret(siret),
		}

		if boardConfig.Board.UserIsActiveMember(user) {
			card.ListID = wekanCard.ListID
			card.BoardID = wekanCard.BoardID
			card.CreatorID = wekanCard.UserID
			card.AssigneeIDs = wekanCard.Assignees
			card.SwimlaneID = wekanCard.SwimlaneID
			card.LabelIDs = wekanCard.LabelIDs
			card.MemberIDs = wekanCard.Members
			card.Description = wekanCard.Description
			card.URL = wekanCardURL(wekanCard)
			card.UserIsBoardMember = true
		}

		return card
	}
}

func (service wekanService) ExportCardsFromSiret(ctx context.Context, siret string, username libwekan.Username) ([]core.KanbanCard, error) {
	pipeline := wekan.BuildCardFromCustomTextFieldPipeline("SIRET", siret)
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

	pipeline = append(pipeline, replaceDateLastActivityStage, removeLastActivityStage)

	cards, err := wekan.SelectCardsFromPipeline(ctx, "customFields", pipeline)
	if err != nil {
		return nil, err
	}
	return utils.Convert(cards, wekanCardToKanbanCard(username)), nil
}
