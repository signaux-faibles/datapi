with actions as (
  select id_campaign_etablissement,
         last(action order by id) as action
  from campaign_etablissement_action
  group by id_campaign_etablissement
), inserted as (
  insert into campaign_etablissement_action (id_campaign_etablissement, username, date_action, action)
    select ce.id, $3, current_timestamp, 'take'
    from campaign_etablissement ce
           inner join v_summaries s on s.siret = ce.siret and code_departement = any ($4)
           left join etablissement_follow f on f.siret = ce.siret and f.username = $3 and f.active
           left join actions a on a.id_campaign_etablissement = ce.id
    where ce.id_campaign = $1
      and ce.id = $2
      and coalesce(a.action, 'cancel') in ('cancel', 'withdraw')
    returning id, id_campaign_etablissement)
select i.id, s.code_departement from inserted i
inner join campaign_etablissement ce on ce.id = i.id_campaign_etablissement
inner join v_summaries s on s.siret = ce.siret
