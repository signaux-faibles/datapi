SELECT vs.siret
FROM v_summaries vs
INNER JOIN codefi_entreprises ce ON vs.siren = ce.siren
WHERE vs.liste = $1
  AND vs.alert != 'Pas d''alerte'
  AND vs.last_procol IN ('in_bonis', 'continuation', 'sauvegarde', 'plan_sauvegarde')
  AND vs.etat_administratif != 'F'
  AND vs.first_alert;
