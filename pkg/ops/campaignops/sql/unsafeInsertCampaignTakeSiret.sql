insert into campaign_etablissement_action (id_campaign_etablissement, username, date_action, action)
select id, $1, current_timestamp, 'take' from campaign_etablissement
where
  siret = $2
and id_campaign = $3
