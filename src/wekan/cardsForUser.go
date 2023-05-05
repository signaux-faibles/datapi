package wekan

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/datapi/src/utils"
	"github.com/signaux-faibles/libwekan"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

type Etablissements struct {
	summaries  core.Summaries
	wekanCards []libwekan.Card
}

func cardToSiret(wc libwekan.Config) func(libwekan.Card) string {
	return func(card libwekan.Card) string {
		siret, _ := wc.GetCardCustomFieldByName(card, "SIRET")
		return siret
	}
}

func cardWithCommentsToSiret(wc libwekan.Config) func(comments libwekan.CardWithComments) string {
	return func(card libwekan.CardWithComments) string {
		siret, _ := wc.GetCardCustomFieldByName(card.Card, "SIRET")
		return siret
	}
}

func selectSummariesWithoutCard(
	ctx context.Context,
	sirets []string,
	db *pgxpool.Pool,
	user libwekan.User,
	roles []string,
	zone []string,
) (core.Summaries, error) {
	rows, err := db.Query(ctx, SqlGetFollow, roles, user.Username, zone, sirets)
	defer rows.Close()
	if err != nil {
		return core.Summaries{}, err
	}

	var summaries core.Summaries
	for rows.Next() {
		s := summaries.NewSummary()
		err := rows.Scan(s...)
		if err != nil {
			return core.Summaries{}, err
		}
	}

	return summaries, nil
}

func selectSummariesWithSirets(
	ctx context.Context,
	sirets []string,
	db *pgxpool.Pool,
	user libwekan.User,
	roles []string,
	zone []string,
	limit int,
) (core.Summaries, error) {
	rows, err := db.Query(ctx, SqlGetCards, roles, user.Username, sirets, zone, limit)
	defer rows.Close()
	if err != nil {
		return core.Summaries{}, err
	}

	var summaries core.Summaries
	for rows.Next() {
		s := summaries.NewSummary()
		err := rows.Scan(s...)
		if err != nil {
			return core.Summaries{}, err
		}
	}

	return summaries, nil
}

func buildCardsForUserPipeline(wc libwekan.Config, params core.KanbanSelectCardsForUserParams) libwekan.Pipeline {
	pipeline := libwekan.Pipeline{}

	if len(params.BoardIDs) > 0 {
		pipeline.AppendStage(buildMatchBoardIDsStage(params.BoardIDs))
	}

	pipeline.AppendPipeline(wekan.BuildDomainCardsPipeline())

	if len(params.Lists) > 0 {
		pipeline.AppendPipeline(buildMatchListsPipeline(params.Lists))
	}

	if params.Type == "my-cards" {
		pipeline.AppendStage(buildMatchUserCardsStage(params.User))
	}

	if len(params.Labels) > 0 {
		pipeline.AppendPipeline(buildMatchLabelsPipeline(params.Labels, params.LabelMode, wc))
	}

	if params.Since != nil {
		pipeline.AppendPipeline(buildMatchLastActivityPipeline(params.Since))
	}

	return pipeline
}

func buildMatchUserCardsStage(user libwekan.User) bson.M {
	return bson.M{
		"$match": bson.M{
			"$expr": bson.M{
				"$or": bson.A{
					bson.M{"$eq": bson.A{"$userId", user.ID}},
					bson.M{"$in": bson.A{user.ID, "$members"}},
					bson.M{"$in": bson.A{user.ID, "$assignees"}},
				},
			},
		},
	}
}

func buildMatchLastActivityPipeline(since *time.Time) libwekan.Pipeline {
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

	matchLastActivityStage := bson.M{
		"$match": bson.M{
			"dateLastActivity": bson.M{"$gte": since},
		},
	}

	return libwekan.Pipeline{
		lookupCommentsStage,
		replaceDateLastActivityStage,
		removeLastActivityStage,
		matchLastActivityStage,
	}
}

func buildMatchListsPipeline(lists []string) libwekan.Pipeline {
	lookupListStage := bson.M{
		"$lookup": bson.M{
			"from":         "lists",
			"as":           "list",
			"localField":   "listId",
			"foreignField": "_id",
		},
	}
	unwindListStage := bson.M{
		"$unwind": "$list",
	}

	matchListTitle := bson.M{
		"$match": bson.M{
			"$expr": bson.M{"$in": bson.A{"$list.title", lists}},
		},
	}

	return libwekan.Pipeline{
		lookupListStage,
		unwindListStage,
		matchListTitle,
	}
}

func buildMatchLabelsPipeline(labels []libwekan.BoardLabelName, labelMode string, wc libwekan.Config) libwekan.Pipeline {
	var pipeline libwekan.Pipeline

	boardIDs := []libwekan.BoardID{}
	for boardID, boardConfig := range wc.Boards {
		if boardConfig.Board.HasAllLabelNames(labels) && labelMode == "and" {
			boardIDs = append(boardIDs, boardID)
		}
		if boardConfig.Board.HasAnyLabelNames(labels) && labelMode == "or" {
			boardIDs = append(boardIDs, boardID)
		}
	}
	pipeline.AppendStage(bson.M{
		"$match": bson.M{
			"$expr": bson.M{
				"$in": bson.A{"$boardId", boardIDs},
			},
		},
	})

	operator := "$and"
	if labelMode == "or" {
		operator = "$or"
	}

	pipeline.AppendStage(bson.M{
		"$match": bson.M{
			"$expr": buildLabelsConditions(labels, operator, wc),
		},
	})

	return pipeline
}

func buildLabelsConditions(labels []libwekan.BoardLabelName, operator string, wc libwekan.Config) bson.M {
	var boardsCondition bson.A
	for boardID, board := range wc.Boards {
		var boardCondition bson.A
		for _, label := range board.Board.Labels {
			if utils.Contains(labels, label.Name) {
				boardCondition = append(boardCondition,
					bson.M{"$in": bson.A{label.ID, "$labelIds"}},
				)
			}
		}
		if boardCondition != nil {
			boardsCondition = append(boardsCondition, bson.M{
				"$and": bson.A{
					bson.M{operator: boardCondition},
					bson.M{"boardId": boardID},
				},
			})
		}
	}
	return bson.M{
		"$or": boardsCondition,
	}
}

func buildMatchBoardIDsStage(boardIDs []libwekan.BoardID) bson.M {
	return bson.M{
		"$match": bson.M{
			"$expr": bson.M{
				"$in": bson.A{
					"$_id", boardIDs,
				},
			},
		},
	}
}

func kanbanDBExportHasSiret(siret string) func(export *core.KanbanDBExport) bool {
	return func(summary *core.KanbanDBExport) bool {
		return summary.Siret == siret
	}
}

func joinKanbanDBExportsWithCards(kanbanDBExports core.KanbanDBExports, cards []libwekan.CardWithComments) core.KanbanExports {
	var kanbanExports core.KanbanExports
	for _, card := range cards {
		siret := cardToSiret(wekanConfig)(card.Card)
		kanbanDBExport, ok := utils.First(kanbanDBExports, kanbanDBExportHasSiret(siret))
		if ok {
			kanbanExports = append(kanbanExports, join(card, *kanbanDBExport))
		}
	}
	return kanbanExports
}

func (w wekanService) SelectFollowsForUser(ctx context.Context, params core.KanbanSelectCardsForUserParams, db *pgxpool.Pool, roles []string) (core.Summaries, error) {
	wc := wekanConfig.Copy()
	pipeline := buildCardsForUserPipeline(wc, params)
	pipeline.AppendStage(bson.M{
		"$project": bson.M{
			"boardId":      true,
			"customFields": true,
		},
	})

	cards, err := wekan.SelectCardsFromPipeline(ctx, "boards", append(bson.A{}, pipeline...))

	if err != nil {
		return core.Summaries{}, err
	}
	sirets := utils.Convert(cards, cardToSiret(wc))
	if params.Type == "my-cards" {
		err := followSiretsFromWekan(ctx, params.User.Username, sirets)
		if err != nil {
			return core.Summaries{}, err
		}
	}
	// my-cards et all-cards utilisent la même méthode
	if utils.Contains([]string{"my-cards", "all-cards"}, params.Type) {
		summaries, err := selectSummariesWithSirets(ctx, sirets, db, params.User, roles, params.Zone, 100)
		return summaries, err
	}
	// no-cards retourne les suivis datapi sans carte kanban
	summaries, err := selectSummariesWithoutCard(ctx, sirets, db, params.User, roles, params.Zone)
	return summaries, err
}
