package campaign

type UnexpectedCampaignError struct {
	err error
}

func (e UnexpectedCampaignError) Error() string {
	return "une erreur s'est produite"
}

func (e UnexpectedCampaignError) Unwrap() error {
	return e.err
}

type CampaignNotFoundError struct {
	err error
}

func (e CampaignNotFoundError) Error() string {
	return "aucune campagne trouvée"
}

func (e CampaignNotFoundError) Unwrap() error {
	return e.err
}
