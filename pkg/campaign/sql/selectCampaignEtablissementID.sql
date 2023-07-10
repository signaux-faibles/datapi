select ce.siret, c.wekan_domain_regexp, s.code_departement, c.id from campaign c
inner join campaign_etablissement ce on ce.id_campaign = c.id
inner join v_summaries s on s.siret = ce.siret
where ce.id = $1
