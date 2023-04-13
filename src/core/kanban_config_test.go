package core

import (
	"github.com/signaux-faibles/datapi/src/test/factory"
	"github.com/signaux-faibles/datapi/src/utils"
	"github.com/signaux-faibles/libwekan"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_kanbanConfigForUser_withActiveMembers(t *testing.T) { // GIVEN
	ass := assert.New(t)
	userOne := factory.OneWekanUser()
	userTwo := factory.OneWekanUser()
	userThree := factory.OneWekanUser()
	configBoardA := factory.OneConfigBoardWithMembers(userOne, userTwo)
	configBoardB := factory.OneConfigBoardWithMembers(userOne, userThree)

	// set up wekanConfig for test
	wekanConfig = factory.LibwekanConfigWith(
		[]libwekan.ConfigBoard{configBoardA, configBoardB},
		[]libwekan.User{userOne, userTwo, userThree},
	)

	// WHEN
	configForUserOne := kanbanConfigForUser(userOne.Username)
	configForUserTwo := kanbanConfigForUser(userTwo.Username)
	configForUserThree := kanbanConfigForUser(userThree.Username)

	// THEN
	ass.Equal(userOne.ID, configForUserOne.UserID)
	configShouldExactlyContains(ass, configForUserOne, configBoardA.Board, configBoardB.Board)

	ass.Equal(userTwo.ID, configForUserTwo.UserID)
	configShouldExactlyContains(ass, configForUserTwo, configBoardA.Board)

	ass.Equal(userThree.ID, configForUserThree.UserID)
	configShouldExactlyContains(ass, configForUserThree, configBoardB.Board)

}

func Test_kanbanConfigForUser_withInactiveMembers(t *testing.T) { // GIVEN
	ass := assert.New(t)
	userOne := factory.OneWekanUser()
	configBoardA := factory.OneConfigBoardWithMembers(userOne)
	factory.DeactiveMembers(configBoardA, userOne)

	// set up wekanConfig for test
	wekanConfig = factory.LibwekanConfigWith(
		[]libwekan.ConfigBoard{configBoardA},
		[]libwekan.User{userOne},
	)

	// WHEN
	configForUserOne := kanbanConfigForUser(userOne.Username)

	// THEN
	ass.Equal(userOne.ID, configForUserOne.UserID)
	ass.Empty(configForUserOne.Boards)
}

func configShouldExactlyContains(ass *assert.Assertions, config KanbanConfig, boards ...libwekan.Board) {
	boardIDsFromConfig := utils.GetKeys(config.Boards)

	expectedBoardIDs := utils.Convert(boards, func(board libwekan.Board) libwekan.BoardID {
		return board.ID
	})
	ass.ElementsMatch(boardIDsFromConfig, expectedBoardIDs)
}
