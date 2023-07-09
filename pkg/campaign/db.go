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

//go:embed sql/selectMyActions.sql
var sqlSelectMyActions string

//go:embed sql/selectTakenActions.sql
var sqlSelectTakenActions string

//go:embed sql/actionMyEtablissement.sql
var sqlActionMyEtablissement string

//go:embed sql/selectCampaignEtablissementID.sql
var sqlSelectCampaignEtablissementID string
