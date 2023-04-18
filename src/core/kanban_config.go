package core

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/signaux-faibles/libwekan"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var wekan libwekan.Wekan
var wekanConfig libwekan.Config
var wekanConfigMutex = &sync.Mutex{}

type KanbanUsers map[libwekan.UserID]libwekan.Username

type KanbanLists map[libwekan.ListID]KanbanList
type KanbanList struct {
	Title string  `json:"title"`
	Sort  float64 `json:"sort"`
}

type KanbanSwimlanes map[libwekan.SwimlaneID]KanbanSwimlane
type KanbanSwimlane struct {
	Title string  `json:"title"`
	Sort  float64 `json:"sort"`
}

type KanbanBoardMembers map[libwekan.UserID]libwekan.Username

type KanbanBoardLabels map[libwekan.BoardLabelID]KanbanBoardLabel
type KanbanBoardLabel struct {
	Color string                  `json:"color"`
	Name  libwekan.BoardLabelName `json:"name"`
}

type KanbanBoards map[libwekan.BoardID]KanbanBoard
type KanbanBoard struct {
	Title     libwekan.BoardTitle `json:"title"`
	Slug      libwekan.BoardSlug  `json:"slug"`
	Lists     KanbanLists         `json:"lists"`
	Swimlanes KanbanSwimlanes     `json:"swimlanes"`
	Labels    KanbanBoardLabels   `json:"labels"`
	Members   KanbanBoardMembers  `json:"members"`
}

type KanbanBoardSwimlane struct {
	BoardID    libwekan.BoardID    `json:"boardID"`
	SwimlaneID libwekan.SwimlaneID `json:"swimlaneID"`
}

type KanbanConfig struct {
	Departements map[CodeDepartement][]KanbanBoardSwimlane `json:"departements"`
	Boards       KanbanBoards                              `json:"boards"`
	Users        KanbanUsers                               `json:"users"`
	UserID       libwekan.UserID                           `json:"userID"`
}

func populateWekanConfigBoards(
	boards map[libwekan.BoardID]libwekan.ConfigBoard,
	wekanUsers map[libwekan.UserID]libwekan.User,
	wekanUser libwekan.User,
) KanbanBoards {
	b := make(KanbanBoards)
	for wekanBoardId, wekanBoard := range boards {
		if wekanBoard.Board.UserIsActiveMember(wekanUser) {
			var kanbanBoard KanbanBoard
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

func fromWekanBoardMembers(
	wekanBoardMembers []libwekan.BoardMember,
	wekanConfigUsers map[libwekan.UserID]libwekan.User) KanbanBoardMembers {
	r := make(KanbanBoardMembers)
	for _, member := range wekanBoardMembers {
		if member.IsActive {
			username := wekanConfigUsers[member.UserID].Username
			r[member.UserID] = username
		}
	}
	return r
}

func fromWekanLists(lists map[libwekan.ListID]libwekan.List) KanbanLists {
	r := make(KanbanLists)
	for listID, wekanList := range lists {
		var kanbanList KanbanList
		kanbanList.Title = wekanList.Title
		kanbanList.Sort = wekanList.Sort
		r[listID] = kanbanList
	}
	return r
}

func fromWekanSwimlanes(swimlanes map[libwekan.SwimlaneID]libwekan.Swimlane) KanbanSwimlanes {
	r := make(KanbanSwimlanes)
	for swimlaneID, wekanSwimlane := range swimlanes {
		var kanbanSwimlane KanbanSwimlane
		kanbanSwimlane.Title = wekanSwimlane.Title
		kanbanSwimlane.Sort = wekanSwimlane.Sort
		r[swimlaneID] = kanbanSwimlane
	}
	return r
}

func fromWekanBoardLabels(labels []libwekan.BoardLabel) KanbanBoardLabels {
	r := make(KanbanBoardLabels)
	for _, wekanBoardLabel := range labels {
		var kanbanBoardLabel KanbanBoardLabel
		kanbanBoardLabel.Color = wekanBoardLabel.Color
		kanbanBoardLabel.Name = wekanBoardLabel.Name
		r[wekanBoardLabel.ID] = kanbanBoardLabel
	}
	return r
}

func (k *KanbanConfig) populateDepartements() {
	kanbanDepartements := make(map[CodeDepartement][]KanbanBoardSwimlane)
	for boardID, board := range k.Boards {
		for swimlaneID, swimlane := range board.Swimlanes {
			zone := CodeDepartement(parseSwimlaneTitle(swimlane))
			if _, ok := departements[zone]; ok {
				current := KanbanBoardSwimlane{
					BoardID:    boardID,
					SwimlaneID: swimlaneID,
				}
				kanbanDepartements[zone] = append(kanbanDepartements[zone], current)
			}
			candidateRegion := Region(parseSwimlaneTitle(swimlane))
			if depts, ok := regions[candidateRegion]; ok {
				for _, dept := range depts {
					current := KanbanBoardSwimlane{
						BoardID:    boardID,
						SwimlaneID: swimlaneID,
					}
					kanbanDepartements[dept] = append(kanbanDepartements[dept], current)
				}
			}
		}
	}
	k.Departements = kanbanDepartements
}

func kanbanConfigForUser(username libwekan.Username) KanbanConfig {
	wekanConfigMutex.Lock()
	config := wekanConfig.Copy()
	wekanConfigMutex.Unlock()
	var kanbanConfig KanbanConfig
	for wekanUserID, wekanUser := range config.Users {
		if wekanUser.Username == username {
			kanbanConfig.UserID = wekanUserID
			kanbanConfig.Boards = populateWekanConfigBoards(config.Boards, config.Users, wekanUser)
			kanbanConfig.populateDepartements()
		}
	}
	return kanbanConfig
}

func kanbanConfigHandler(c *gin.Context) {
	var s session
	s.bind(c)
	kanbanConfig := kanbanConfigForUser(libwekan.Username(s.username))
	c.JSON(http.StatusOK, kanbanConfig)
}

func parseSwimlaneTitle(swimlane KanbanSwimlane) string {
	titleSplitted := strings.Split(swimlane.Title, " (")
	if len(titleSplitted) > 2 {
		log.Println("Erreur : le titre de la swimlane est mal configuré : ", swimlane.Title)
	}
	return strings.TrimSpace(titleSplitted[0])
}

func loadWekanConfig() {
	newWekanConfig, err := wekan.SelectConfig(context.Background())
	if err != nil {
		log.Printf("loadWekanConfig() -> ERROR : problem loading config: %s", err.Error())
	} else {
		wekanConfigMutex.Lock()
		wekanConfig = newWekanConfig
		wekanConfigMutex.Unlock()
	}
}

func watchWekanConfig(period time.Duration) {
	for {
		loadWekanConfig()
		time.Sleep(period)
	}
}
