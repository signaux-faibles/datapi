truncate table entreprise_bce;

delete
from etablissement
where siren not in (select siren
                    from score
                    where batch = '2309'
                      and alert != 'Pas d''alerte');


delete
from entreprise
where siren not in (select siren
                    from score
                    where batch = '2309'
                      and alert != 'Pas d''alerte');


delete
from score
where siren not in (select siren
                    from score
                    where batch = '2309'
                      and alert != 'Pas d''alerte');

delete
from entreprise_ellisphere
where siren not in (select siren
                    from score
                    where batch = '2309'
                      and alert != 'Pas d''alerte');

delete
from etablissement_periode_urssaf
where siren not in (select siren
                    from score
                    where batch = '2309'
                      and alert != 'Pas d''alerte');

delete
from etablissement_apconso
where siren not in (select siren
                    from score
                    where batch = '2309'
                      and alert != 'Pas d''alerte');

delete
from etablissement_apdemande
where siren not in (select siren
                    from score
                    where batch = '2309'
                      and alert != 'Pas d''alerte');

delete
from entreprise_ellisphere
where siren not in (select siren
                    from score
                    where batch = '2309'
                      and alert != 'Pas d''alerte');

delete
from entreprise_paydex
where siren not in (select siren
                    from score
                    where batch = '2309'
                      and alert != 'Pas d''alerte');

delete
from etablissement_delai
where siren not in (select siren
                    from score
                    where batch = '2309'
                      and alert != 'Pas d''alerte');

delete
from entreprise_pge
where siren not in (select siren
                    from score
                    where batch = '2309'
                      and alert != 'Pas d''alerte');

delete
from etablissement_procol
where siren not in (select siren
                    from score
                    where batch = '2309'
                      and alert != 'Pas d''alerte');

delete from etablissement_comments;
delete from etablissement_follow;
delete
from entreprise_ellisphere
where siren not in (select siren
                    from score
                    where batch = '2309'
                      and alert != 'Pas d''alerte');

delete
from entreprise_diane
where siren not in (select siren
                    from score
                    where batch = '2309'
                      and alert != 'Pas d''alerte');



refresh materialized view v_alert_entreprise;
refresh materialized view v_alert_etablissement;
refresh materialized view v_apdemande;
refresh materialized view v_diane_variation_ca;
refresh materialized view v_etablissement_raison_sociale;
refresh materialized view v_hausse_urssaf;
refresh materialized view v_last_effectif;
refresh materialized view v_last_procol;
refresh materialized view v_naf;
refresh materialized view v_roles;
refresh materialized view v_summaries;
