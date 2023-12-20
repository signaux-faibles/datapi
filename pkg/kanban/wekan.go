package kanban

import (
	"context"
	"datapi/pkg/core"
	"fmt"
	"github.com/signaux-faibles/libwekan"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"strings"
	"sync"
	"time"
)

var wekan libwekan.Wekan
var WekanConfig libwekan.Config
var wekanConfigMutex = &sync.Mutex{}

type wekanService struct{}

func (service wekanService) LoadConfigForUser(username libwekan.Username) core.KanbanConfig {
	return kanbanConfigForUser(username)
}

func (service wekanService) GetUser(username libwekan.Username) (libwekan.User, bool) {
	return GetUser(username)
}

func (service wekanService) SelectCardsFromSiret(ctx context.Context, siret string, username libwekan.Username) ([]core.KanbanCard, error) {
	return selectCardsFromSiret(ctx, siret, username)
}

func GetUser(username libwekan.Username) (libwekan.User, bool) {
	wekanConfigMutex.Lock()
	defer wekanConfigMutex.Unlock()
	return WekanConfig.GetUserByUsername(username)
}

func InitService(ctx context.Context, dBURL, dBName, admin, slugDomainRegexp string) core.KanbanService {
	var err error
	wekan, err = libwekan.Init(ctx, dBURL, dBName, libwekan.Username(admin), slugDomainRegexp)
	if err != nil {
		log.Printf("Erreur lors de l'initialisation de wekan : %s", err)
	}
	go watchWekanConfig(time.Minute)
	return wekanService{}
}

func kanbanConfigForUser(username libwekan.Username) core.KanbanConfig {
	wekanConfigMutex.Lock()
	config := WekanConfig.Copy()
	wekanConfigMutex.Unlock()
	var kanbanConfig core.KanbanConfig
	for wekanUserID, wekanUser := range config.Users {
		if wekanUser.Username == username {
			kanbanConfig.UserID = wekanUserID
			kanbanConfig.Boards = populateWekanConfigBoards(config.Boards, config.Users, wekanUser)
			kanbanConfig.Departements = populateDepartements(kanbanConfig)
			kanbanConfig.Users = populateUsers(wekanUserID, config.Boards, config.Users)
		}
	}
	return kanbanConfig
}

func populateUsers(
	currentUserID libwekan.UserID,
	boards map[libwekan.BoardID]libwekan.ConfigBoard,
	users map[libwekan.UserID]libwekan.User,
) core.KanbanUsers {
	r := make(core.KanbanUsers)
	for _, value := range boards {
		for _, member := range value.Board.Members {
			user, found := users[member.UserID]
			if found {
				user := core.KanbanUser{
					Username: user.Username,
					Active:   member.IsActive,
					Fullname: users[member.UserID].Profile.Fullname,
				}
				r[member.UserID] = user

			}
		}
	}
	return r
}

func populateWekanConfigBoards(
	boards map[libwekan.BoardID]libwekan.ConfigBoard,
	wekanUsers map[libwekan.UserID]libwekan.User,
	wekanUser libwekan.User,
) core.KanbanBoards {
	b := make(core.KanbanBoards)
	for wekanBoardId, wekanBoard := range boards {
		if wekanBoard.Board.UserIsActiveMember(wekanUser) {
			var kanbanBoard core.KanbanBoard
			kanbanBoard.Title = wekanBoard.Board.Title
			kanbanBoard.Slug = wekanBoard.Board.Slug
			kanbanBoard.Lists = fromWekanLists(wekanBoard.Lists)
			kanbanBoard.Swimlanes = fromWekanSwimlanes(wekanBoard.Swimlanes)
			kanbanBoard.Labels = fromWekanBoardLabels(wekanBoard.Board.Labels)
			kanbanBoard.Members = fromWekanBoardMembers(wekanBoard.Board.Members, wekanUsers)
			b[wekanBoardId] = kanbanBoard
		}
	}
	return b
}

func populateDepartements(kanbanConfig core.KanbanConfig) map[core.CodeDepartement][]core.KanbanBoardSwimlane {
	kanbanDepartements := make(map[core.CodeDepartement][]core.KanbanBoardSwimlane)
	for boardID, board := range kanbanConfig.Boards {
		for swimlaneID, swimlane := range board.Swimlanes {
			zone := core.CodeDepartement(parseSwimlaneTitle(swimlane))
			if _, ok := core.Departements[zone]; ok {
				current := core.KanbanBoardSwimlane{
					BoardID:    boardID,
					SwimlaneID: swimlaneID,
				}
				kanbanDepartements[zone] = append(kanbanDepartements[zone], current)
			}
			candidateRegion := core.Region(parseSwimlaneTitle(swimlane))
			if depts, ok := core.Regions[candidateRegion]; ok {
				for _, dept := range depts {
					current := core.KanbanBoardSwimlane{
						BoardID:    boardID,
						SwimlaneID: swimlaneID,
					}
					kanbanDepartements[dept] = append(kanbanDepartements[dept], current)
				}
			}
		}
	}
	return kanbanDepartements
}

func fromWekanBoardMembers(
	wekanBoardMembers []libwekan.BoardMember,
	wekanConfigUsers map[libwekan.UserID]libwekan.User) core.KanbanBoardMembers {
	r := make(core.KanbanBoardMembers)
	for _, member := range wekanBoardMembers {
		r[member.UserID] = core.KanbanBoardMember{
			UserID: member.UserID,
			Active: member.IsActive,
		}
	}
	return r
}

func fromWekanLists(lists map[libwekan.ListID]libwekan.List) core.KanbanLists {
	r := make(core.KanbanLists)
	for listID, wekanList := range lists {
		var kanbanList core.KanbanList
		kanbanList.Title = wekanList.Title
		kanbanList.Sort = wekanList.Sort
		r[listID] = kanbanList
	}
	return r
}

func fromWekanSwimlanes(swimlanes map[libwekan.SwimlaneID]libwekan.Swimlane) core.KanbanSwimlanes {
	r := make(core.KanbanSwimlanes)
	for swimlaneID, wekanSwimlane := range swimlanes {
		var kanbanSwimlane core.KanbanSwimlane
		kanbanSwimlane.Title = wekanSwimlane.Title
		kanbanSwimlane.Sort = wekanSwimlane.Sort
		r[swimlaneID] = kanbanSwimlane
	}
	return r
}

func fromWekanBoardLabels(labels []libwekan.BoardLabel) core.KanbanBoardLabels {
	r := make(core.KanbanBoardLabels)
	for _, wekanBoardLabel := range labels {
		var kanbanBoardLabel core.KanbanBoardLabel
		kanbanBoardLabel.Color = wekanBoardLabel.Color
		kanbanBoardLabel.Name = wekanBoardLabel.Name
		r[wekanBoardLabel.ID] = kanbanBoardLabel
	}
	return r
}

func parseSwimlaneTitle(swimlane core.KanbanSwimlane) string {
	titleSplitted := strings.Split(swimlane.Title, " (")
	if len(titleSplitted) > 2 {
		log.Println("Erreur : le titre de la swimlane est mal configuré : ", swimlane.Title)
	}
	return strings.TrimSpace(titleSplitted[0])
}

func (service wekanService) SelectSiretsFromListeAndDomainRegexp(ctx context.Context, wekanDomainRegexp string, liste string) ([]core.Siret, error) {
	pipeline := buildCardsFromListeAndDomainRegexpPipeline(wekanDomainRegexp, liste)
	cards, err := wekan.SelectCardsFromPipeline(ctx, "boards", pipeline)
	if err != nil {
		return nil, err
	}
	var sirets []core.Siret
	for _, card := range cards {
		fmt.Println("coucou card")
		if siret, ok := WekanConfig.GetCardCustomFieldByName(card, "SIRET"); ok {
			fmt.Println("coucou")
			sirets = append(sirets, core.Siret(siret))
		}
	}
	return sirets, nil
}

func buildCardsFromListeAndDomainRegexpPipeline(wekanDomainRegexp string, liste string) libwekan.Pipeline {
	var pipeline libwekan.Pipeline
	pipeline.AppendStage(bson.M{
		"$match": bson.M{"slug": primitive.Regex{Pattern: wekanDomainRegexp, Options: "i"}},
	})
	pipeline.AppendStage(bson.M{
		"$lookup": bson.M{
			"from":         "cards",
			"localField":   "_id",
			"foreignField": "boardId",
			"as":           "card",
		}})
	pipeline.AppendStage(bson.M{
		"$unwind": "$card",
	})
	pipeline.AppendStage(bson.M{
		"$lookup": bson.M{
			"from":         "lists",
			"localField":   "card.listId",
			"foreignField": "_id",
			"as":           "card.list",
		}})
	pipeline.AppendStage(bson.M{
		"$unwind": "$card.list",
	})
	pipeline.AppendStage(bson.M{
		"$match": bson.M{
			"card.list.title": "Accompagnement en cours",
		}})
	pipeline.AppendStage(bson.M{
		"$replaceRoot": bson.M{
			"newRoot": "$card",
		},
	})
	return pipeline
}

// [

//    {$unwind: '$card.list'},
//    {$match: {
//        'card.list.title': 'Accompagnement en cours',
//    }},
//    {$lookup: {
//        from: 'customFields',
//        localField: 'card.listId',
//        foreignField: '_id',
//        as: 'card.list'
//    }},
//    {$unwind:
//        '$card.customFields'
//    },
//    {$match: {
//        'card.customFields.value': /^[0-9]{14}/
//    }},
//    {$project: {
//        _id: 0,
//        siret: '$card.customFields.value'
//    }}
//]
