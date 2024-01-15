select siret
from v_summaries
where last_list=$1 and alert != 'Pas d''alerte'
  and last_procol in ('in_bonis', 'continuation', 'sauvegarde', 'plan_sauvegarde')
  and etat_administratif != 'F'
  and first_alert;