package factory

import (
	"github.com/gosimple/slug"
	"github.com/jaswdr/faker"
	"github.com/signaux-faibles/datapi/src/utils"
	"github.com/signaux-faibles/libwekan"
)

var fake faker.Faker

func init() {
	fake = faker.New()
}

// OneWekanUserWithID methode factory
func OneWekanUserWithID(userID string) libwekan.User {
	user := libwekan.User{
		ID:       libwekan.UserID(userID),
		Username: libwekan.Username(fake.Internet().SafeEmail()),
		IsAdmin:  fake.Bool(),
	}
	return user
}

// OneWekanUser methode factory
func OneWekanUser() libwekan.User {
	return OneWekanUserWithID(fake.UUID().V4())
}

// OneConfigBoardWithActiveMembers methode factory
func OneConfigBoardWithActiveMembers(member ...libwekan.User) libwekan.ConfigBoard {
	boardMembers := utils.Convert(member, func(user libwekan.User) libwekan.BoardMember {
		return libwekan.BoardMember{
			UserID:   user.ID,
			IsActive: true,
			IsAdmin:  fake.Bool(),
		}
	})
	title := libwekan.BoardTitle(fake.Address().Country())
	slitle := libwekan.BoardSlug(slug.Make(string(title)))
	board := libwekan.Board{
		ID:               libwekan.BoardID(fake.UUID().V4()),
		Title:            title,
		Members:          boardMembers,
		Slug:             slitle,
		Watchers:         nil,
		AllowsCardNumber: false,
		AllowsShowLists:  false,
	}
	r := libwekan.ConfigBoard{
		Board:        board,
		Swimlanes:    nil,
		Lists:        nil,
		CustomFields: nil,
	}
	return r
}

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
