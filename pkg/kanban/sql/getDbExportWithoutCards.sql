select v.siret, v.raison_sociale, v.code_departement, v.libelle_departement, v.commune,
       coalesce(v.code_territoire_industrie, ''), coalesce(v.libelle_territoire_industrie, ''), v.siege, coalesce(v.raison_sociale_groupe, ''),
       coalesce(v.code_activite, ''), coalesce(v.libelle_n5, 'norme NAF non prise en charge'), coalesce(v.libelle_n1, 'norme NAF non prise en charge'),
       v.statut_juridique_n1, v.statut_juridique_n2, v.statut_juridique_n3, coalesce(v.date_ouverture_etablissement, '1900-01-01'),
       coalesce(v.date_creation_entreprise, '1900-01-01'), coalesce(v.effectif_entreprise, 0), coalesce(v.date_effectif, '1900-01-01'),
       coalesce(v.arrete_bilan, '0001-01-01'), coalesce(v.exercice_diane,0), coalesce(v.chiffre_affaire,0),
       coalesce(v.prev_chiffre_affaire,0), coalesce(v.variation_ca, 1), coalesce(v.resultat_expl,0),
       coalesce(v.prev_resultat_expl,0), coalesce(v.excedent_brut_d_exploitation,0),
       coalesce(v.prev_excedent_brut_d_exploitation,0), coalesce(v.last_list,''), coalesce(v.last_alert,''),
       coalesce(v.activite_partielle, false), coalesce(v.hausse_urssaf, false), coalesce(v.presence_part_salariale, false),
       coalesce(v.periode_urssaf, '0001-01-01'), v.last_procol, coalesce(v.date_last_procol, '0001-01-01'), coalesce(f.since, '0001-01-01'),
       coalesce(f.comment, ''), (permissions($1, v.roles, v.first_list_entreprise, v.code_departement, f.siret is not null)).in_zone
from v_summaries v
       inner join etablissement_follow f on f.siret = v.siret and f.username = $2 and active
where (v.code_departement = any($4) or coalesce($4, '{}') = '{}') and (not (v.siret = any($3)) or $3 is null) and (v.raison_sociale ilike $5 or $5 is null)
order by f.id, v.siret
