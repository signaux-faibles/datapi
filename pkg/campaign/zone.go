package campaign

import (
	"datapi/pkg/core"
	"datapi/pkg/utils"
	"github.com/signaux-faibles/libwekan"
	"strings"
)

type BoardZones map[string][]string
type Zone []string

func zonesFromBoards(boards []libwekan.ConfigBoard) BoardZones {
	zones := make(map[string][]string)
	for _, board := range boards {
		swimlanes := utils.GetValues(board.Swimlanes)
		zone := utils.Convert(swimlanes, func(s libwekan.Swimlane) string {
			return strings.Split(string(s.Title), " (")[0]
		})
		zones[string(board.Board.Slug)] = zone
	}
	return zones
}

func zoneForUser(username string) Zone {
	boards := core.Kanban.SelectBoardsForUsername(libwekan.Username(username))
	zones := zonesFromBoards(boards)
	return zoneFromBoardZones(zones)
}

func zoneFromBoardZones(zones map[string][]string) Zone {
	var zone []string
	for _, zs := range zones {
		for _, z := range zs {
			zone = append(zone, z)
		}
	}
	return utils.Uniq(zone)
}
