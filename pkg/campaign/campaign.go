package campaign

import (
	"time"
)

//func campaignsForUser(userBoards core.KanbanBoards, allCampaigns Campaigns) Campaigns {
//	var userCampaigns = make(Campaigns, 0)
//	for _, campaign := range allCampaigns {
//		if campaign.matchesBoards(utils.GetValues(userBoards)) {
//			userCampaigns = append(userCampaigns, campaign)
//		}
//	}
//	return userCampaigns
//}
//
//func (campaign Campaign) matchesBoards(configBoards []core.KanbanBoard) bool {
//	re, err := regexp2.Compile(campaign.WekanDomainRegexp)
//	if err != nil {
//		return false
//	}
//	return utils.Any(configBoards, func(kanbanBoard core.KanbanBoard) bool {
//		return re.MatchString(string(kanbanBoard.Slug))
//	})
//}

type Campaign struct {
	ID                int       `json:"id"`
	Libelle           string    `json:"libelle"`
	DateEnd           time.Time `json:"date_fin"`
	DateCreate        time.Time `json:"date_create"`
	WekanDomainRegexp string    `json:"wekan_domain_regexp"`
	BoardSlugs        []string  `json:"slugs"`
	NBTotal           int       `json:"nb_total"`
	NBPerimetre       int       `json:"nb_perimetre"`
}

type Campaigns []*Campaign

func (cs *Campaigns) Tuple() []interface{} {
	c := Campaign{}
	*cs = append(*cs, &c)
	return []interface{}{
		&c.ID,
		&c.Libelle,
		&c.WekanDomainRegexp,
		&c.DateEnd,
		&c.DateCreate,
		&c.BoardSlugs,
		&c.NBTotal,
		&c.NBPerimetre,
	}
}
