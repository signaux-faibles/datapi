with zones as (
    select key as slug, ARRAY(SELECT jsonb_array_elements_text(value)) as zone
    from jsonb_each($1::jsonb)
  ), id_boards as (
    select key as slug, value->>0 as id_board from jsonb_each($2::jsonb)
), campaigns as (select c.id,
      c.libelle,
      c.wekan_domain_regexp,
      date_end,
      date_create,
      array_agg(z.slug) as slugs, flatmap(z.zone) as zones, array_agg(i.id_board) as id_boards
   from campaign c
   inner join zones z on z.slug ~ c.wekan_domain_regexp
   inner join id_boards i on i.slug = z.slug
   group by c.id, c.libelle, c.wekan_domain_regexp, date_end, date_create)
select cs.id, cs.libelle, cs.wekan_domain_regexp, cs.date_end, cs.date_create,
       coalesce(count(distinct ce.id), 0) as nb_total,
       coalesce(sum(case when s.code_departement = any(cs.zones) then 1 else 0 end), 0) as nb_perimetre,
       cs.zones, cs.id_boards
from campaigns cs
left join campaign_etablissement ce on ce.id_campaign = cs.id
left join v_summaries s on s.siret = ce.siret
group by cs.id, cs.libelle, cs.wekan_domain_regexp, cs.date_end, cs.date_create, cs.slugs, cs.zones, cs.id_boards
