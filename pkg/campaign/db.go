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

//go:embed sql/withdrawPendingEtablissement.sql
var sqlWithdrawPendingEtablissement string

//go:embed sql/checkSirets.sql
var sqlCheckSirets string

//go:embed sql/addSirets.sql
var sqlAddSirets string

//go:embed sql/selectExports.sql
var sqlSelectExports string
