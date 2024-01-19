update score s set siret = v.siret
from v_summaries v
where v.siren = s.siren and v.siege
and s.batch = $1