package campaign

import (
	"datapi/pkg/core"
	"datapi/pkg/utils"
	regexp2 "regexp"
	"time"
)

func campaignsForUser(userBoards core.KanbanBoards, allCampaigns Campaigns) Campaigns {
	var userCampaigns Campaigns
	for _, campaign := range allCampaigns {
		if campaign.matchesBoards(utils.GetValues(userBoards)) {
			userCampaigns = append(userCampaigns, campaign)
		}
	}
	return userCampaigns
}

func (campaign Campaign) matchesBoards(configBoards []core.KanbanBoard) bool {
	re, err := regexp2.Compile(campaign.WekanDomainRegexp)
	if err != nil {
		return false
	}
	return utils.Any(configBoards, func(kanbanBoard core.KanbanBoard) bool {
		return re.MatchString(string(kanbanBoard.Slug))
	})
}

type Campaign struct {
	ID                int       `json:"id"`
	Libelle           string    `json:"libelle"`
	DateEnd           time.Time `json:"date_fin"`
	DateCreate        time.Time `json:"date_create"`
	WekanDomainRegexp string    `json:"wekan_domain_regexp"`
}

type Campaigns []*Campaign

func (cs *Campaigns) NewRowItems() []interface{} {
	c := Campaign{}
	*cs = append(*cs, &c)
	return []interface{}{
		&c.ID,
		&c.Libelle,
		&c.WekanDomainRegexp,
		&c.DateEnd,
		&c.DateCreate,
	}
}
