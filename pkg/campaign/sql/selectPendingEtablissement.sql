with actions as (
  select id_campaign_etablissement,
    last(action order by id) as action
  from campaign_etablissement_action
  group by id_campaign_etablissement
), zones as (
  select key as slug, ARRAY(SELECT jsonb_array_elements_text(value)) as zone
  from jsonb_each($2::jsonb)
), zone as (
  select flatmap(z.zone) as zone
  from campaign c
  inner join zones z on z.slug ~ c.wekan_domain_regexp
  where id = $1
)
select count(*) over () as nb_total,
   s.siret,
   s.raison_sociale,
   s.raison_sociale_groupe,
   s.alert,
   ce.id,
   ce.id_campaign,
   f.id is not null as followed,
   s.first_alert,
   s.etat_administratif,
   a.action,
   rank() over (order by ce.id) as rank,
   s.code_departement
from campaign_etablissement ce
  inner join zone z on true
  inner join v_summaries s on s.siret = ce.siret and s.code_departement = any(z.zone)
  left join etablissement_follow f on f.siret = ce.siret and f.username=$3 and f.active
  left join actions a on a.id_campaign_etablissement = ce.id
where ce.id_campaign = $1 and coalesce(a.action, 'cancel') = 'cancel'


