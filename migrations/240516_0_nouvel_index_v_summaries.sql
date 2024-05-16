drop index idx_v_summaries_score;

create index if not exists idx_v_summaries_score on
public.v_summaries
	using btree (
  valeur_score desc, 
  siret, 
  siege,
  effectif, 
  effectif_entreprise, 
  code_departement, 
  last_procol, 
  chiffre_affaire, 
  etat_administratif, 
  first_alert, 
  first_list_entreprise, 
  first_red_list_entreprise, 
  has_delai, 
  date_creation_entreprise,
  code_activite)
where
(alert <> 'Pas d''alerte'::text)