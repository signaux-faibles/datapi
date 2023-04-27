package core

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/signaux-faibles/libwekan"
	"time"
)

// KanbanService service définissant les méthodes de Kanban nécessaires dans Datapi
// TODO il  faudrait changer les types `libwekan` en des types `datapi`
type KanbanService interface {
	LoadConfigForUser(username libwekan.Username) KanbanConfig
	SelectCardsFromSiret(ctx context.Context, siret string, username libwekan.Username) ([]KanbanCard, error)
	SelectCardsForUser(ctx context.Context, params KanbanSelectCardsForUserParams, db *pgxpool.Pool, roles []string) ([]*Summary, error)
	GetUser(username libwekan.Username) (libwekan.User, bool)
	CreateCard(ctx context.Context, params KanbanNewCardParams, username libwekan.Username) error
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

type KanbanCard struct {
	ID                libwekan.CardID         `json:"id,omitempty"`
	ListID            libwekan.ListID         `json:"listID,omitempty"`
	ListTitle         string                  `json:"listTitle,omitempty"`
	Archived          bool                    `json:"archived"`
	BoardID           libwekan.BoardID        `json:"boardID,omitempty"`
	BoardTitle        libwekan.BoardTitle     `json:"boardTitle,omitempty"`
	SwimlaneID        libwekan.SwimlaneID     `json:"swimlaneID,omitempty"`
	URL               string                  `json:"url,omitempty"`
	Description       string                  `json:"description,omitempty"`
	AssigneeIDs       []libwekan.UserID       `json:"assigneesID,omitempty"`
	MemberIDs         []libwekan.UserID       `json:"memberIDs,omitempty"`
	CreatorID         libwekan.UserID         `json:"creatorID,omitempty"`
	Creator           libwekan.Username       `json:"creator,omitempty"`
	LastActivity      time.Time               `json:"lastActivity,omitempty"`
	StartAt           time.Time               `json:"startAt,omitempty"`
	EndAt             *time.Time              `json:"endAt,omitempty"`
	LabelIDs          []libwekan.BoardLabelID `json:"labelIds,omitempty"`
	UserIsBoardMember bool                    `json:"userIsBoardMember"`
}

type KanbanSelectCardsForUserParams struct {
	User      libwekan.User             `json:"-"`
	Type      string                    `json:"type"`
	Zone      []string                  `json:"zone"`
	BoardIDs  []libwekan.BoardID        `json:"boardIDs"`
	Labels    []libwekan.BoardLabelName `json:"labels"`
	Since     *time.Time                `json:"since"`
	Lists     []string                  `json:"lists"`
	LabelMode string                    `json:"labelMode"`
}

type KanbanFollows struct {
	Count   int       `json:"count"`
	Follows []Summary `json:"summaries"`
}

type KanbanNewCardParams struct {
	SwimlaneID  libwekan.SwimlaneID     `json:"swimlaneID"`
	Description string                  `json:"description"`
	LabelIDs    []libwekan.BoardLabelID `json:"labelIDs"`
	Siret       Siret                   `json:"siret"`
}
