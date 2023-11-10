package campaign

import (
	"context"
)

type CampaignEffect interface {
	Do(ctx context.Context) error
}

func buildEffect(effect string, id CampaignEtablissementID) CampaignEffect {
	switch effect {
	case "follow":
		{
			return CampaignFollowEffect{
				id: id,
			}
		}
	}
	return CampaignNilEffect{}
}

type CampaignNilEffect struct{}

func (c CampaignNilEffect) Do(ctx context.Context) error {
	return nil
}

type CampaignFollowEffect struct {
	id CampaignEtablissementID
}

func (c CampaignFollowEffect) Do(ctx context.Context) error {
	// On ajoute le participant à la carte
	// Il en découle un accompagnement en cours
	return nil
}
