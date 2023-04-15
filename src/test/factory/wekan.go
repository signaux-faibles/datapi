package factory

import (
	"github.com/gosimple/slug"
	"github.com/signaux-faibles/datapi/src/utils"
	"github.com/signaux-faibles/libwekan"
)

// OneWekanUser methode factory
func OneWekanUser() libwekan.User {
	person := fake.Person()
	user := libwekan.BuildUser(person.Contact().Email, person.Title(), person.Name())
	return user.Admin(fake.Bool())
}

// OneConfigBoardWithMembers methode factory
func OneConfigBoardWithMembers(members ...libwekan.User) libwekan.ConfigBoard {
	boardMembers := utils.Convert(members, func(user libwekan.User) libwekan.BoardMember {
		return libwekan.BoardMember{
			UserID:   user.ID,
			IsActive: true,
			IsAdmin:  fake.Bool(),
		}
	})
	title := fake.Beer().Name()
	slitle := slug.Make(title)
	board := libwekan.BuildBoard(title, slitle, fake.Beer().Style())
	board.Members = boardMembers
	r := libwekan.ConfigBoard{
		Board:        board,
		Swimlanes:    nil,
		Lists:        nil,
		CustomFields: nil,
	}
	return r
}

// DeactiveMembers désactive les membres pour une configuration de Board
func DeactiveMembers(config libwekan.ConfigBoard, members ...libwekan.User) {
	memberIds := utils.Convert(members, func(m libwekan.User) libwekan.UserID {
		return m.ID
	})
	boardMembers := config.Board.Members
	for index, member := range boardMembers {
		if utils.Contains(memberIds, member.UserID) {
			// on desactive l'utilisateur courant via son adresse mémoire
			(&boardMembers[index]).IsActive = false
		}
	}
}

// LibwekanConfigWith construit une `libwekan.Config`
func LibwekanConfigWith(configBoards []libwekan.ConfigBoard, users []libwekan.User) libwekan.Config {
	boards := utils.ToMap(configBoards, func(board libwekan.ConfigBoard) libwekan.BoardID {
		return board.Board.ID
	})
	u := utils.ToMap(users, func(user libwekan.User) libwekan.UserID {
		return user.ID
	})
	r := libwekan.Config{
		Boards: boards,
		Users:  u,
	}
	return r
}

func AddSwimlanesWithDepartments(config *libwekan.ConfigBoard, codes ...string) map[libwekan.SwimlaneID]libwekan.Swimlane {
	r := make(map[libwekan.SwimlaneID]libwekan.Swimlane)
	for _, code := range codes {
		title := code + " (" + fake.Color().SafeColorName() + ")"
		order := fake.Float64(1000, 0, 999)
		swimlane := libwekan.BuildSwimlane(config.Board.ID, fake.App().Name(), title, order)
		r[swimlane.ID] = swimlane
	}
	config.Swimlanes = r
	return r
}
