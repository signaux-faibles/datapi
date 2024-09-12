truncate table v_summaries;
insert into v_summaries
WITH last_liste AS (SELECT first(liste.libelle ORDER BY liste.batch DESC, liste.algo DESC) AS last_liste,
                           first('20' || substring(batch from 1 for 2) || '-' || substring(batch from 3 for 2) || '-01'
                                 ORDER BY liste.batch DESC, liste.algo DESC
                           )::date                                                         as date_last_liste
                    FROM liste),
     entreprises_avec_delai as (select distinct siren
                                from etablissement_delai
                                       join last_liste l on true
                                where stade in ('APPROB', 'PR')
                                  and date_echeance > l.date_last_liste)
SELECT r.roles,
       et.siret,
       et.siren,
       en.raison_sociale,
       et.commune,
       d.libelle                                                                                           AS libelle_departement,
       et.departement                                                                                      AS code_departement,
       s.score                                                                                             AS valeur_score,
       s.detail                                                                                            AS detail_score,
       COALESCE(aet.first_list = l.last_liste OR aet.first_red_list = l.last_liste,
                false)                                                                                     AS first_alert,
       aen.first_list                                                                                      AS first_list_entreprise,
       aen.first_red_list                                                                                  AS first_red_list_entreprise,
       aet.first_list                                                                                      AS first_list_etablissement,
       aet.first_red_list                                                                                  AS first_red_list_etablissement,
       aet.last_list,
       aet.last_alert,
       l.last_liste                                                                                        AS liste,
       di.chiffre_affaire,
       di.prev_chiffre_affaire,
       di.arrete_bilan,
       di.exercice_diane,
       di.variation_ca,
       di.resultat_expl,
       di.prev_resultat_expl,
       di.excedent_brut_d_exploitation,
       di.prev_excedent_brut_d_exploitation,
       ef.effectif,
       sum(ef.effectif)
       FILTER (WHERE COALESCE(et.etat_administratif, 'A'::text) = 'A'::text) OVER (PARTITION BY en.siren)  AS effectif_entreprise,
       max(ef.periode)
       FILTER (WHERE COALESCE(et.etat_administratif, 'A'::text) <> 'A'::text) OVER (PARTITION BY en.siren) AS date_entreprise,
       ef.periode                                                                                          AS date_effectif,
       n.libelle_n5,
       n.libelle_n1,
       et.code_activite,
       COALESCE(ep.last_procol, 'in_bonis'::text)                                                          AS last_procol,
       ep.date_effet                                                                                       AS date_last_procol,
       COALESCE(ap.ap, false)                                                                              AS activite_partielle,
       ap.heure_consomme                                                                                   AS apconso_heure_consomme,
       ap.montant                                                                                          AS apconso_montant,
       u.dette[1] > u.dette[2] OR u.dette[2] > u.dette[3]                                                  AS hausse_urssaf,
       u.dette[1]                                                                                          AS dette_urssaf,
       u.periode_urssaf,
       u.presence_part_salariale,
       s.alert,
       et.siege,
       g.raison_sociale                                                                                    AS raison_sociale_groupe,
       ti.code_commune IS NOT NULL                                                                         AS territoire_industrie,
       ti.code_terrind                                                                                     AS code_territoire_industrie,
       ti.libelle_terrind                                                                                  AS libelle_territoire_industrie,
       cj.libelle                                                                                          AS statut_juridique_n3,
       cj2.libelle                                                                                         AS statut_juridique_n2,
       cj1.libelle                                                                                         AS statut_juridique_n1,
       et.creation                                                                                         AS date_ouverture_etablissement,
       en.creation                                                                                         AS date_creation_entreprise,
       COALESCE(sco.secteur_covid, 'nonSecteurCovid'::text)                                                AS secteur_covid,
       CASE
         WHEN COALESCE(en.etat_administratif, 'A'::text) = 'C'::text THEN 'F'::text
         ELSE COALESCE(et.etat_administratif, 'A'::text)
         END                                                                                               AS etat_administratif,
       en.etat_administratif                                                                               AS etat_administratif_entreprise,
       ed.siren is not null                                                                                as has_delai
FROM last_liste l
       JOIN etablissement0 et ON true
       JOIN entreprise0 en ON en.siren::text = et.siren::text
       left JOIN entreprises_avec_delai ed on en.siren = ed.siren
       JOIN v_etablissement_raison_sociale etrs ON etrs.id_etablissement = et.id
       JOIN v_roles r ON et.siren::text = r.siren::text
       JOIN departements d ON d.code = et.departement
       LEFT JOIN v_naf n ON n.code_n5 = et.code_activite
       LEFT JOIN secteurs_covid sco ON sco.code_activite = et.code_activite
       LEFT JOIN score0 s ON et.siret::text = s.siret AND s.libelle_liste = l.last_liste
       LEFT JOIN v_alert_etablissement aet ON aet.siret = et.siret::text
       LEFT JOIN v_alert_entreprise aen ON aen.siren = et.siren::text
       LEFT JOIN v_last_effectif ef ON ef.siret::text = et.siret::text
       LEFT JOIN v_hausse_urssaf u ON u.siret::text = et.siret::text
       LEFT JOIN v_apdemande ap ON ap.siret::text = et.siret::text
       LEFT JOIN v_last_procol ep ON ep.siren::text = et.siren::text
       LEFT JOIN v_diane_variation_ca di ON di.siren::text = et.siren::text
       LEFT JOIN entreprise_ellisphere0 g ON g.siren::text = et.siren::text
       LEFT JOIN terrind ti ON ti.code_commune = et.code_commune
       LEFT JOIN categorie_juridique cj ON cj.code = en.statut_juridique
       LEFT JOIN categorie_juridique cj2 ON "substring"(cj.code, 0, 3) = cj2.code
       LEFT JOIN categorie_juridique cj1 ON "substring"(cj.code, 0, 2) = cj1.code;

select count(*)
from v_summaries
