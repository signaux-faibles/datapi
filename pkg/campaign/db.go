package campaign

import (
	_ "embed"
)

//go:embed sql/selectMatchingCampaigns.sql
var sqlSelectMatchingCampaigns string

//go:embed sql/selectPendingEtablissement.sql
var sqlSelectPendingEtablissement string

//go:embed sql/takePendingEtablissement.sql
var sqlTakePendingEtablissement string
