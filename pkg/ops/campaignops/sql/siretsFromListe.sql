select siret
from v_summaries
where last_list=$1 and alert != 'Pas d''alerte'
  and last_procol in ('in_bonis', 'redressement', 'plan_redressement')
  and etat_administratif != 'F'
  and first_alert;