package campaign

type CampaignEtablissementActionID int

func (c CampaignEtablissementActionID) Tuple() []interface{} {
	return []interface{}{&c}
}
