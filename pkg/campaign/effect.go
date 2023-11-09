package campaign

import (
	"context"
	"fmt"
)

type CampaignEffect interface {
	Do(ctx context.Context) error
}

func buildEffect(effect string, id CampaignEtablissementID) CampaignEffect {
	switch effect {
	case "FOLLOW":
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
	fmt.Println("Je fais le suivi %d", c.id)
	return nil
}
