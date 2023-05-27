package kanban

import (
	"context"
	"github.com/signaux-faibles/libwekan"
	"log"
	"time"
)

func watchWekanConfig(period time.Duration) {
	for {
		loadWekanConfig()
		time.Sleep(period)
	}
}

func loadWekanConfig() {
	newWekanConfig, err := wekan.SelectConfig(context.Background())
	if err != nil {
		log.Printf("Erreur lors du chargement de la config wekan pour l'utilisateur : %s", err)
	} else {
		wekanConfigMutex.Lock()
		WekanConfig = newWekanConfig
		wekanConfigMutex.Unlock()
	}
}

func wekanConfigForUser(wc libwekan.Config, user libwekan.User) libwekan.Config {
	new := libwekan.Config{
		Boards: make(map[libwekan.BoardID]libwekan.ConfigBoard),
	}

	wekanConfigMutex.Lock()
	old := wc.Copy()
	wekanConfigMutex.Unlock()

	for boardID, board := range old.Boards {
		if board.Board.UserIsActiveMember(user) {
			new.Boards[boardID] = board
		}
	}

	return new
}

func (service wekanService) ClearBoardIDs(boardIDs []libwekan.BoardID, user libwekan.User) []libwekan.BoardID {
	wc := WekanConfig.Copy()
	var newBoardIDs []libwekan.BoardID
	if len(boardIDs) == 0 {
		for id, board := range wc.Boards {
			if board.Board.UserIsActiveMember(user) {
				newBoardIDs = append(newBoardIDs, id)
			}
		}
	} else {
		for _, id := range boardIDs {
			if board, ok := wc.Boards[id]; ok {
				if board.Board.UserIsActiveMember(user) {
					newBoardIDs = append(newBoardIDs, id)
				}
			}
		}
	}
	return newBoardIDs
}
		wekanConfig = newWekanConfig
		wekanConfigMutex.Unlock()
	}
}
