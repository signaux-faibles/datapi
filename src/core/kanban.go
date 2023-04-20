package core

import (
	"github.com/signaux-faibles/libwekan"
)

// KanbanService service définissant les méthodes de Kanban nécessaires dans Datapi
// TODO il  faudrait changer les types `libwekan` en des types `datapi`
type KanbanService interface {
	LoadConfigForUser(username libwekan.Username) KanbanConfig
	GetUser(username libwekan.Username) (libwekan.User, bool)
}

type KanbanUsers map[libwekan.UserID]KanbanUser

type KanbanUser struct {
	Username libwekan.Username `json:"username"`
	Active   bool              `json:"active"`
}

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

type KanbanBoardMembers map[libwekan.UserID]KanbanBoardMember
type KanbanBoardMember struct {
	Username libwekan.Username `json:"username"`
	Active   bool              `json:"active"`
}

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
