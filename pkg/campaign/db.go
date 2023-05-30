package campaign

import (
	"context"
	"datapi/pkg/db"
	"github.com/signaux-faibles/libwekan"
)

func selectMatchingCampaigns(ctx context.Context, boardSlugs []libwekan.BoardSlug, roles []string) (Campaigns, error) {
	sql := `with boards as (
      select unnest($1::text[]) slug
    ), campaigns as (select c.id,
        c.libelle,
        c.wekan_domain_regexp,
        date_end,
        date_create,
        array_agg(b.slug order by b.slug) slugs
     from campaign c
        inner join boards b on b.slug ~ c.wekan_domain_regexp
     group by c.id, c.libelle, c.wekan_domain_regexp, date_end, date_create)
    select cs.id, cs.libelle, cs.wekan_domain_regexp, cs.date_end, cs.date_create, cs.slugs,
           count(distinct ce.id) as nb_total,
           sum(case when s.code_departement = any($2) then 1 else 0 end) as nb_perimetre
    from campaigns cs
    inner join campaign_etablissement ce on ce.id_campaign = cs.id
    inner join v_summaries s on s.siret = ce.siret
    group by cs.id, cs.libelle, cs.wekan_domain_regexp, cs.date_end, cs.date_create, cs.slugs`

	var allCampaigns = Campaigns{}
	err := db.Query(ctx, &allCampaigns, sql, boardSlugs, roles)

	return allCampaigns, err
}
