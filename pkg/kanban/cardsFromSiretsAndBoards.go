package kanban

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/utils"
	"fmt"
	"github.com/signaux-faibles/libwekan"
	"go.mongodb.org/mongo-driver/bson"
)

func (service wekanService) SelectCardsFromSiretsAndBoardIDs(ctx context.Context, sirets []core.Siret, boardIDs []libwekan.BoardID, username libwekan.Username) ([]core.KanbanCard, error) {
	pipeline := wekan.BuildCardFromCustomTextFieldsPipeline(
		"SIRET",
		utils.Convert(sirets, func(siret core.Siret) string { return string(siret) }),
	)

	matchBoardIDsStage := bson.M{
		"$match": bson.M{"boardId": bson.M{"$in": boardIDs}},
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

	pipeline = append(pipeline, matchBoardIDsStage, replaceDateLastActivityStage, removeLastActivityStage)

	cards, err := wekan.SelectCardsFromPipeline(ctx, "customFields", pipeline)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return utils.Convert(cards, wekanCardToKanbanCard(username)), nil
}
