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

type PendingNotFoundError struct {
	err error
}

func (e PendingNotFoundError) Error() string {
	return "établissement indisponible pour ce traitement"
}

func (e PendingNotFoundError) Unwrap() error {
	return e.err
}

type TakeNotFoundError struct {
	err error
}

func (e TakeNotFoundError) Error() string {
	return "établissement indisponible pour ce traitement"
}

func (e TakeNotFoundError) Unwrap() error {
	return e.err
}

type InvalidCampaignDomainError struct {
	domain string
	err    error
}

func (e InvalidCampaignDomainError) Error() string {
	return "domaine wekan de la campagne invalide: " + e.domain
}

func (e InvalidCampaignDomainError) Unwrap() error {
	return e.err
}
