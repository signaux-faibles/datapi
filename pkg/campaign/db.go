package campaign

import (
	"context"
	"datapi/pkg/db"
)

func selectAllCampaignsFromDB(ctx context.Context) (Campaigns, error) {
	sql := "select id, libelle, wekan_domain_regexp, date_end, date_create " +
		"from campaign"

	var allCampaigns = Campaigns{}
	err := db.Query(ctx, &allCampaigns, sql)

	return allCampaigns, err
}
