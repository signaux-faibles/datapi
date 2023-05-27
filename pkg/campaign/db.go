package campaign

import (
	"context"
	"datapi/pkg/db"
	"github.com/signaux-faibles/libwekan"
)

func selectMatchingCampaigns(ctx context.Context, boardSlugs []libwekan.BoardSlug) (Campaigns, error) {
	sql := `with boards as (
      select unnest($1::text[]) slug
    )
    select c.id, c.libelle, c.wekan_domain_regexp, date_end, date_create,
    array_agg(b.slug order by b.slug)
    from campaign c
    inner join boards b on b.slug ~ c.wekan_domain_regexp
    group by c.id, c.libelle, c.wekan_domain_regexp, date_end, date_create`

	var allCampaigns = Campaigns{}
	err := db.Query(ctx, &allCampaigns, sql, boardSlugs)

	return allCampaigns, err
}
