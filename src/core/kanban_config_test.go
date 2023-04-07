package core

import (
	"bou.ke/monkey"
	"github.com/signaux-faibles/datapi/src/factory"
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
	configBoardA := factory.OneConfigBoardWithActiveMembers(userOne, userTwo)
	configBoardB := factory.OneConfigBoardWithActiveMembers(userOne, userThree)

	wekanConfig = factory.LibwekanConfigWith(
		[]libwekan.ConfigBoard{configBoardA, configBoardB},
		[]libwekan.User{userOne, userTwo, userThree},
	)
	//MockLibwekanConfig(wekanConfig, mockedConfig)

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

//func Test_kanbanConfigForUser_withInactiveMembers(t *testing.T) {
//
//	// GIVEN
//	ass := assert.New(t)
//	userOne := factory.OneWekanUser()
//	configBoardA := factory.OneConfigBoardWithInactiveMembers(userOne)
//	factory.DeactiveUsersOnBoard(configBoardA, userOne)
//
//	mockedConfig := factory.LibwekanConfigWith(
//		[]libwekan.ConfigBoard{configBoardA},
//		[]libwekan.User{userOne},
//	)
//	MockLibwekanConfig(mockedConfig)
//
//	// WHEN
//	configForUserOne := kanbanConfigForUser(userOne.Username)
//
//	// THEN
//	ass.Equal(userOne.ID, configForUserOne.UserID)
//	ass.Empty(configForUserOne.Boards)
//
//}

func configShouldExactlyContains(ass *assert.Assertions, config KanbanConfig, boards ...libwekan.Board) {
	boardIDsFromConfig := utils.GetKeys(config.Boards)
	//boardIDsFromConfig := make([]libwekan.BoardID, 0, len(config.Boards))
	//for key := range config.Boards {
	//	boardIDsFromConfig = append(boardIDsFromConfig, key)
	//}

	expectedBoardIDs := utils.Convert(boards, func(board libwekan.Board) libwekan.BoardID {
		return board.ID
	})
	ass.ElementsMatch(boardIDsFromConfig, expectedBoardIDs)
}

// MockLibwekanConfig méthode qui permet mocker la config wekan
func MockLibwekanConfig(original, mock libwekan.Config) {
	monkey.Patch((&original).Copy, func(*libwekan.Config) libwekan.Config {
		return mock
	})
}
