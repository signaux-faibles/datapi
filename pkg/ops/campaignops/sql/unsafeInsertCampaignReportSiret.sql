insert into campaign_etablissement_action (id_campaign_etablissement, username, date_action, action, detail)
select id, $1, current_timestamp, 'withdraw', $2 from campaign_etablissement
where
siret = $3
and id_campaign = $4
