truncate table v_roles;
insert into v_roles
SELECT etablissement.siren,
       array_agg(DISTINCT etablissement.departement) AS roles
FROM etablissement
WHERE etablissement.version = 0
GROUP BY etablissement.siren;


truncate table v_naf;
insert into v_naf
SELECT n1.id      AS id_n1,
       n2.id      AS id_n2,
       n3.id      AS id_n3,
       n4.id      AS in_n4,
       n5.id      AS id_n5,
       n1.code    AS code_n1,
       n2.code    AS code_n2,
       n3.code    AS code_n3,
       n4.code    AS code_n4,
       n5.code    AS code_n5,
       n1.libelle AS libelle_n1,
       n2.libelle AS libelle_n2,
       n3.libelle AS libelle_n3,
       n4.libelle AS libelle_n4,
       n5.libelle AS libelle_n5
FROM naf n5
       JOIN naf n4 ON n5.id_parent = n4.id
       JOIN naf n3 ON n4.id_parent = n3.id
       JOIN naf n2 ON n3.id_parent = n2.id
       JOIN naf n1 ON n2.id_parent = n1.id;

truncate table v_last_procol;
insert into v_last_procol
SELECT etablissement_procol0.siren,
       CASE
         WHEN 'liquidation'::text = ANY (array_agg(etablissement_procol0.action_procol)) THEN 'liquidation'::text
         WHEN 'redressement'::text = ANY (array_agg(etablissement_procol0.action_procol)) THEN
           CASE
             WHEN 'redressement_plan_continuation'::text = ANY
                  (array_agg((etablissement_procol0.action_procol || '_'::text) ||
                             etablissement_procol0.stade_procol)) THEN 'plan_continuation'::text
             ELSE 'redressement'::text
             END
         WHEN 'sauvegarde'::text = ANY (array_agg(etablissement_procol0.action_procol)) THEN
           CASE
             WHEN 'sauvegarde_plan_continuation'::text = ANY
                  (array_agg((etablissement_procol0.action_procol || '_'::text) ||
                             etablissement_procol0.stade_procol)) THEN 'plan_sauvegarde'::text
             ELSE 'sauvegarde'::text
             END
         ELSE NULL::text
         END                                 AS last_procol,
       max(etablissement_procol0.date_effet) AS date_effet
FROM etablissement_procol0
GROUP BY etablissement_procol0.siren;

truncate table v_last_effectif;
insert into v_last_effectif
WITH last_effectif_entreprise AS (SELECT etablissement_periode_urssaf0.siren,
                                         last(etablissement_periode_urssaf0.periode
                                              ORDER BY etablissement_periode_urssaf0.periode) AS periode
                                  FROM etablissement_periode_urssaf0
                                  WHERE etablissement_periode_urssaf0.effectif IS NOT NULL
                                  GROUP BY etablissement_periode_urssaf0.siren)
SELECT e.siret,
       e.siren,
       e.effectif,
       e.periode
FROM etablissement_periode_urssaf0 e
       JOIN last_effectif_entreprise l ON e.siren::text = l.siren::text AND e.periode = l.periode;

truncate table v_hausse_urssaf;
insert into v_hausse_urssaf
SELECT etablissement_periode_urssaf0.siret,
       (array_agg(etablissement_periode_urssaf0.part_patronale + etablissement_periode_urssaf0.part_salariale
                  ORDER BY etablissement_periode_urssaf0.periode DESC))[0:3] AS dette,
       CASE
         WHEN first(etablissement_periode_urssaf0.part_salariale
                    ORDER BY etablissement_periode_urssaf0.periode DESC) > 100::double precision THEN true
         ELSE false
         END                                                                 AS presence_part_salariale,
       max(etablissement_periode_urssaf0.periode)                            AS periode_urssaf
FROM etablissement_periode_urssaf0
GROUP BY etablissement_periode_urssaf0.siret;

create unique index if not exists idx_v_hausse_urssaf_siret
  on v_hausse_urssaf (siret);

truncate table v_etablissement_raison_sociale;
insert into v_etablissement_raison_sociale
SELECT et.id AS id_etablissement,
       en.id AS id_entreprise,
       en.raison_sociale,
       et.siret
FROM entreprise0 en
       JOIN etablissement0 et ON en.siren::text = et.siren::text;

truncate table v_diane_variation_ca;
insert into v_diane_variation_ca
WITH diane AS (SELECT entreprise_diane0.siren,
                      entreprise_diane0.arrete_bilan_diane,
                      entreprise_diane0.exercice_diane::integer                                                 AS exercice_diane,
                      lag(entreprise_diane0.arrete_bilan_diane, 1)
                      OVER (PARTITION BY entreprise_diane0.siren ORDER BY entreprise_diane0.arrete_bilan_diane) AS prev_arrete_bilan_diane,
                      entreprise_diane0.chiffre_affaire,
                      lag(entreprise_diane0.chiffre_affaire, 1)
                      OVER (PARTITION BY entreprise_diane0.siren ORDER BY entreprise_diane0.arrete_bilan_diane) AS prev_chiffre_affaire,
                      entreprise_diane0.resultat_expl,
                      lag(entreprise_diane0.resultat_expl, 1)
                      OVER (PARTITION BY entreprise_diane0.siren ORDER BY entreprise_diane0.arrete_bilan_diane) AS prev_resultat_expl,
                      entreprise_diane0.excedent_brut_d_exploitation,
                      lag(entreprise_diane0.excedent_brut_d_exploitation, 1)
                      OVER (PARTITION BY entreprise_diane0.siren ORDER BY entreprise_diane0.arrete_bilan_diane) AS prev_excedent_brut_d_exploitation
               FROM entreprise_diane0
               ORDER BY entreprise_diane0.siren, entreprise_diane0.arrete_bilan_diane DESC)
SELECT diane.siren,
       first(diane.arrete_bilan_diane ORDER BY diane.arrete_bilan_diane DESC)   AS arrete_bilan,
       first(diane.exercice_diane ORDER BY diane.arrete_bilan_diane DESC)       AS exercice_diane,
       first(diane.prev_chiffre_affaire ORDER BY diane.arrete_bilan_diane DESC) AS prev_chiffre_affaire,
       first(diane.chiffre_affaire ORDER BY diane.arrete_bilan_diane DESC)      AS chiffre_affaire,
       first(
         CASE
           WHEN diane.prev_chiffre_affaire <> 0::double precision
             THEN diane.chiffre_affaire / diane.prev_chiffre_affaire
           ELSE NULL::real
           END ORDER BY diane.arrete_bilan_diane DESC)                          AS variation_ca,
       first(diane.resultat_expl ORDER BY diane.arrete_bilan_diane DESC)        AS resultat_expl,
       first(diane.prev_resultat_expl ORDER BY diane.arrete_bilan_diane DESC)   AS prev_resultat_expl,
       first(diane.excedent_brut_d_exploitation
             ORDER BY diane.arrete_bilan_diane DESC)                            AS excedent_brut_d_exploitation,
       first(diane.prev_excedent_brut_d_exploitation
             ORDER BY diane.arrete_bilan_diane DESC)                            AS prev_excedent_brut_d_exploitation
FROM diane
GROUP BY diane.siren;

truncate table v_apdemande;
insert into v_apdemande
SELECT e.siret,
       COALESCE((SELECT DISTINCT true AS bool
                 FROM etablissement_apdemande0
                 WHERE etablissement_apdemande0.siret::text = e.siret::text
                   AND (etablissement_apdemande0.periode_end + '1 year'::interval) >=
                       date_trunc('month'::text, CURRENT_TIMESTAMP)), false)      AS ap,
       (sum(c.heure_consomme ORDER BY c.periode) / 12::double precision)::integer AS heure_consomme,
       (sum(c.montant ORDER BY c.periode) / 12::double precision)::integer        AS montant
FROM etablissement0 e
       LEFT JOIN etablissement_apconso0 c ON c.siret::text = e.siret::text AND (c.periode + '1 year'::interval) >=
                                                                               date_trunc('month'::text, CURRENT_TIMESTAMP)
GROUP BY e.siret;

truncate table v_alert_etablissement;
insert into v_alert_etablissement
WITH lists AS (SELECT score0.siret,
                      first(score0.libelle_liste ORDER BY score0.batch) AS first_list,
                      last(score0.libelle_liste ORDER BY score0.batch)  AS last_list,
                      first(score0.alert ORDER BY score0.batch DESC)    AS last_alert
               FROM score0
               WHERE score0.alert = ANY (ARRAY ['Alerte seuil F1'::text, 'Alerte seuil F2'::text])
               GROUP BY score0.siret),
     red_lists AS (SELECT score0.siret,
                          first(score0.libelle_liste ORDER BY score0.batch) AS first_red_list
                   FROM score0
                   WHERE score0.alert = 'Alerte seuil F1'::text
                   GROUP BY score0.siret)
SELECT l.siret,
       l.first_list,
       l.last_list,
       l.last_alert,
       rl.first_red_list
FROM lists l
       LEFT JOIN red_lists rl ON rl.siret = l.siret;

truncate table v_alert_entreprise;
insert into v_alert_entreprise
WITH first_lists AS (SELECT DISTINCT sc.siren,
                                     first(sc.libelle_liste) OVER (window_siren)                           AS first_list,
                                     first(sc.libelle_liste)
                                     FILTER (WHERE sc.alert = 'Alerte seuil F1'::text) OVER (window_siren) AS first_red_list
                     FROM score0 sc
                     WHERE sc.alert = ANY (ARRAY ['Alerte seuil F1'::text, 'Alerte seuil F2'::text])
                     WINDOW window_siren AS (PARTITION BY sc.siren ORDER BY sc.batch))
SELECT first_lists.siren,
       max(first_lists.first_list)     AS first_list,
       max(first_lists.first_red_list) AS first_red_list
FROM first_lists
WHERE first_lists.first_list IS NOT NULL
GROUP BY first_lists.siren;

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
