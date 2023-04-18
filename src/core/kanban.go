package core

import (
	"context"
	"sync"

	"github.com/signaux-faibles/libwekan"
	"log"
	"time"
)

var wekan libwekan.Wekan
var WekanConfig libwekan.Config
var WekanConfigMutex = &sync.Mutex{}

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

func loadWekanConfig() {
	newWekanConfig, err := wekan.SelectConfig(context.Background())
	if err != nil {
		log.Printf("loadWekanConfig() -> ERROR : problem loading config: %s", err.Error())
	} else {
		WekanConfigMutex.Lock()
		WekanConfig = newWekanConfig
		WekanConfigMutex.Unlock()
	}
}

func watchWekanConfig(period time.Duration) {
	for {
		loadWekanConfig()
		time.Sleep(period)
	}
}
