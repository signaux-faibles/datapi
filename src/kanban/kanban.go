package kanban

import (
	"github.com/signaux-faibles/datapi/src/core"
	"github.com/signaux-faibles/libwekan"
	"log"
	"strings"
)

func kanbanConfigForUser(username libwekan.Username) core.KanbanConfig {
	core.WekanConfigMutex.Lock()
	config := core.WekanConfig.Copy()
	core.WekanConfigMutex.Unlock()
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
			// on ne rajoute pas le user courant dans les users de la board
			if member.UserID == currentUserID {
				continue
			}
			user, found := users[member.UserID]
			if found {
				r[member.UserID] = user.Username
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

func populateDepartements(k core.KanbanConfig) map[core.CodeDepartement][]core.KanbanBoardSwimlane {
	kanbanDepartements := make(map[core.CodeDepartement][]core.KanbanBoardSwimlane)
	for boardID, board := range k.Boards {
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
		if member.IsActive {
			username := wekanConfigUsers[member.UserID].Username
			r[member.UserID] = username
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
