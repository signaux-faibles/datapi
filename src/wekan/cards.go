package wekan

import (
	"context"
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/datapi/src/utils"
	"github.com/signaux-faibles/libwekan"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
)

func selectCardsFromSiret(ctx context.Context, siret string, username libwekan.Username) ([]core.KanbanCard, error) {
	pipeline := libwekan.BuildCardFromCustomTextFieldPipeline("SIRET", siret)

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

//// TODO libwekan
//func selectWekanCards(
//	ctx context.Context,
//	username libwekan.Username,
//	boardIds []libwekan.BoardID,
//	swimlanes []string,
//	lists []string,
//	labelIds []libwekan.BoardLabelName,
//	labelMode bool,
//	since *time.Time,
//) ([]libwekan.Card, error) {
//	pipeline := cardsPipeline(username, boardIds, swimlaneIds, listIds, labelIds, labelMode, since)
//
//	cards, err := wekan.SelectCardsFromPipeline(ctx, "cards", pipeline)
//	return cards, err
//}
//
//func cardsPipeline(
//	username libwekan.Username,
//	boardIds []libwekan.BoardID,
//	swimlaneIds []libwekan.SwimlaneID,
//	listIds []libwekan.ListID,
//	labelIds [][]libwekan.BoardLabelID,
//	labelMode bool,
//	since *time.Time,
//) bson.A {
//	query := bson.M{
//		"type": "cardType-card",
//	}
//
//	if username != nil {
//		userID, ok := wekanConfig.GetUserByUsername(username)
//		query["$expr"] = bson.M{
//			"$or": bson.A{
//				bson.M{"$in": bson.A{userID, "$members"}},
//				bson.M{"$in": bson.A{userID, "$assignees"}},
//			},
//		}
//	}
//
//	if len(boardIds) > 0 { // rejet des ids non connus dans la configuration
//		var ids []string
//		for _, id := range boardIds {
//			oldWekanConfig.mu.Lock()
//			if _, ok := oldWekanConfig.BoardIds[id]; ok {
//				ids = append(ids, id)
//			}
//			oldWekanConfig.mu.Unlock()
//		}
//		query["boardId"] = bson.M{"$in": ids}
//	} else { // limiter au périmètre des boards de la configuration
//		query["boardId"] = bson.M{"$in": oldWekanConfig.boardIds()}
//	}
//
//	if len(swimlaneIds) > 0 {
//		query["swimlaneId"] = bson.M{"$in": swimlaneIds}
//	}
//
//	if len(listIds) > 0 {
//		query["listId"] = bson.M{"$in": listIds}
//	}
//
//	if len(labelIds) > 0 {
//		labelQueryAnd := bson.A{}
//		for _, l := range labelIds {
//			labelQueryOr := bson.A{}
//			for _, i := range l {
//				labelQueryOr = append(labelQueryOr, bson.M{
//					"labelIds": i.labelID,
//					"boardId":  i.boardID,
//				})
//			}
//			labelQueryAnd = append(labelQueryAnd, bson.M{
//				"$or": labelQueryOr,
//			})
//		}
//		if labelMode {
//			query["$and"] = labelQueryAnd
//		} else {
//			query["$or"] = labelQueryAnd
//		}
//	}
//
//	pipeline := bson.A{
//		bson.M{
//			"$match": query,
//		},
//		bson.M{
//			"$lookup": bson.M{
//				"from": "card_comments",
//				"let":  bson.M{"card": "$_id"},
//				"pipeline": bson.A{
//					bson.M{
//						"$match": bson.M{
//							"$expr": bson.M{"$eq": bson.A{"$cardId", "$$card"}},
//							"text":  bson.M{"$regex": "#export"},
//						},
//					}, bson.M{
//						"$sort": bson.M{
//							"createdAt": -1,
//						},
//					}, bson.M{
//						"$project": bson.M{
//							"text": 1,
//							"_id":  0,
//						},
//					},
//				},
//				"as": "comments",
//			},
//		},
//		bson.M{
//			"$lookup": bson.M{
//				"from":         "card_comments",
//				"localField":   "_id",
//				"foreignField": "cardId",
//				"as":           "lastActivity",
//			},
//		},
//		bson.M{
//			"$sort": bson.D{
//				bson.E{Key: "archived", Value: 1},
//				bson.E{Key: "startAt", Value: 1},
//			},
//		},
//		bson.M{
//			"$project": bson.M{
//				"archived":     1,
//				"title":        1,
//				"listId":       1,
//				"boardId":      1,
//				"members":      1,
//				"swimlaneId":   1,
//				"customFields": 1,
//				"sort":         1,
//				"labelIds":     1,
//				"startAt":      1,
//				"endAt":        1,
//				"description":  1,
//				"assignees":    1,
//				"comments":     "$comments.text",
//				"lastActivity": bson.M{
//					"$max": bson.A{
//						bson.M{"$max": "$lastActivity.modifiedAt"},
//						"$dateLastActivity",
//					},
//				},
//			},
//		},
//	}
//
//	if since != nil {
//		pipeline = append(pipeline, bson.M{
//			"$match": bson.M{
//				"lastActivity": bson.M{"$gte": since},
//			},
//		})
//	}
//	return pipeline
//}
