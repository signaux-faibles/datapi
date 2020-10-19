alter table score add detail jsonb;

create view v_summary as
select 
  et.siret,
  et.siren,
  en.raison_sociale,
  et.commune,
  d.libelle as libelle_departement,
  d.code as code_departement,
  s.batch as batch_liste,
  s.algo as algo_liste,
  s.libelle_liste,
  s.alert,
  aen.first_list as first_list_entreprise,
  aet.first_list as first_list_etablissement,
  s.score,
  s.detail as detail_score,
  di.chiffre_affaire,
  di.arrete_bilan,
  di.variation_ca,
  di.resultat_expl,
  ef.effectif,
  n.libelle_n5,
  n.libelle_n1,
  et.code_activite,
  coalesce(ep.last_procol, 'in_bonis') as last_procol,
  coalesce(ap.ap, false) activite_partielle,
  u.dette[0] > u.dette[1] or u.dette[1] > u.dette[2] as hausse_urssaf,
  et.siege,
  g.raison_sociale as raison_sociale_groupe,
  ti.code_commune is not null as territoire_industrie
from etablissement0 et
  inner join v_roles r on et.siren = r.siren
  inner join entreprise0 en on en.siren = r.siren
  inner join departements d on d.code = et.departement
  left join v_naf n on n.code_n5 = et.code_activite
  left join score s on et.siret = s.siret
  left join v_alert_etablissement aet on aet.siret = et.siret
  left join v_alert_entreprise aen on aen.siren = et.siren
  left join v_last_effectif ef on ef.siret = et.siret
  left join v_hausse_urssaf u on u.siret = et.siret
  left join v_apdemande ap on ap.siret = et.siret
  left join v_last_procol ep on ep.siret = et.siret
  left join v_diane_variation_ca di on di.siren = s.siren
  left join entreprise_ellisphere0 g on g.siren = et.siren
  left join terrind ti on ti.code_commune = et.code_commune;