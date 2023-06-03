with zones as (
    select key as slug, ARRAY(SELECT jsonb_array_elements_text(value)) as zone
    from jsonb_each($1::jsonb)
  ), campaigns as (select c.id,
      c.libelle,
      c.wekan_domain_regexp,
      date_end,
      date_create,
      array_agg(z.slug) as slugs, flatmap(z.zone) as zones
   from campaign c
   inner join zones z on z.slug ~ c.wekan_domain_regexp
   group by c.id, c.libelle, c.wekan_domain_regexp, date_end, date_create)
select cs.id, cs.libelle, cs.wekan_domain_regexp, cs.date_end, cs.date_create, cs.slugs,
       coalesce(count(distinct ce.id), 0) as nb_total,
       coalesce(sum(case when s.code_departement = any(cs.zones) then 1 else 0 end), 0) as nb_perimetre
from campaigns cs
left join campaign_etablissement ce on ce.id_campaign = cs.id
left join v_summaries s on s.siret = ce.siret
group by cs.id, cs.libelle, cs.wekan_domain_regexp, cs.date_end, cs.date_create, cs.slugs
