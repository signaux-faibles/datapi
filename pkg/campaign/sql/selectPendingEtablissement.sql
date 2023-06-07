with actions as (
  select id_campaign_etablissement,
    last(action order by id) as action
  from campaign_etablissement_action
  group by id_campaign_etablissement
)
select count(*) over () as nb_total,
   s.siret,
   s.raison_sociale,
   s.raison_sociale_groupe,
   s.alert,
   ce.id,
   f.id is not null as followed,
   s.first_alert,
   s.etat_administratif,
   a.action,
   rank() over (order by ce.id) as rank
from campaign_etablissement ce
  inner join v_summaries s on s.siret = ce.siret and s.code_departement = any($4)
  left join etablissement_follow f on f.siret = ce.siret and f.username='christophe.ninucci@beta.gouv.fr' and f.active
  left join actions a on a.id_campaign_etablissement = ce.id
where ce.id_campaign = $3 and coalesce(a.action, 'cancel') = 'cancel'
limit $2::integer offset ($1::integer)*($2::integer)

