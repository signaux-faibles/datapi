package campaign

import (
	"context"
	"datapi/pkg/core"
	"datapi/pkg/db"
)

type CampaignEtablissement struct {
	Siret                       core.Siret `json:"siret"`
	Alert                       string     `json:"alert"`
	Visible                     bool       `json:"visible"`
	InZone                      bool       `json:"inZone"`
	Followed                    bool       `json:"followed"`
	FollowedEntreprise          bool       `json:"followedEntreprise"`
	FirstAlert                  bool       `json:"firstAlert"`
	EtatAdministratif           string     `json:"etatAdministratif"`
	EtatAdministratifEntreprise string     `json:"etatAdministratifEntreprise"`
	State                       string     `json:"campainEtablissementActions"`
	Rank                        int        `json:"rank"`
}

type Pending struct {
	Etablissements []*CampaignEtablissement `json:"etablissements"`
	NbTotal        int                      `json:"nbTotal"`
	Page           int                      `json:"page"`
	PageMax        int                      `json:"pageMax"`
	PageSize       int                      `json:"pageSize"`
}

func (p *Pending) Tuple() []interface{} {
	var ce CampaignEtablissement
	p.Etablissements = append(p.Etablissements, &ce)
	return []interface{}{
		&p.NbTotal,
		&p.Page,
		&p.PageMax,
		&p.PageSize,
		&ce.Siret,
		&ce.Alert,
		&ce.Visible,
		&ce.InZone,
		&ce.Followed,
		&ce.FollowedEntreprise,
		&ce.FirstAlert,
		&ce.EtatAdministratif,
		&ce.EtatAdministratifEntreprise,
		&ce.State,
		&ce.Rank,
	}
}

func selectPending(ctx context.Context, campaignID CampaignID, zone []string, page core.Page) (pending Pending, err error) {
	err = db.Scan(ctx, &pending, sqlSelectPendingEtablissement, campaignID, zone, page.Size, page.From())
	return
}
