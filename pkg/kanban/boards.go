package kanban

import (
	"datapi/pkg/utils"
	"github.com/signaux-faibles/libwekan"
)

func (ws wekanService) SelectBoardsForUsername(username libwekan.Username) []libwekan.ConfigBoard {
	return SelectBoardsForUser(username)
}

func SelectBoardsForUser(username libwekan.Username) []libwekan.ConfigBoard {
	user, ok := wekanConfig.GetUserByUsername(username)
	if !ok {
		return nil
	}
	configBoards := utils.GetValues(wekanConfig.Boards)
	userBoards := utils.Filter(configBoards, userIsBoardActiveMember(user))
	return userBoards
}

func userIsBoardActiveMember(user libwekan.User) func(board libwekan.ConfigBoard) bool {
	return func(board libwekan.ConfigBoard) bool {
		return board.Board.UserIsActiveMember(user)
	}
}
