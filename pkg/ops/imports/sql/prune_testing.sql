truncate table entreprise_bce;

delete
from etablissement
where siren not in (select siren
                    from score
                    where batch = '2312'
                      and alert != 'Pas d''alerte');


delete
from entreprise
where siren not in (select siren
                    from score
                    where batch = '2312'
                      and alert != 'Pas d''alerte');


delete
from score
where siren not in (select siren
                    from score
                    where batch = '2312'
                      and alert != 'Pas d''alerte');

delete
from entreprise_ellisphere
where siren not in (select siren
                    from score
                    where batch = '2312'
                      and alert != 'Pas d''alerte');

delete
from etablissement_periode_urssaf
where siren not in (select siren
                    from score
                    where batch = '2312'
                      and alert != 'Pas d''alerte');

delete
from etablissement_apconso
where siren not in (select siren
                    from score
                    where batch = '2312'
                      and alert != 'Pas d''alerte');

delete
from etablissement_apdemande
where siren not in (select siren
                    from score
                    where batch = '2312'
                      and alert != 'Pas d''alerte');

delete
from entreprise_ellisphere
where siren not in (select siren
                    from score
                    where batch = '2312'
                      and alert != 'Pas d''alerte');

delete
from entreprise_paydex
where siren not in (select siren
                    from score
                    where batch = '2312'
                      and alert != 'Pas d''alerte');

delete
from etablissement_delai
where siren not in (select siren
                    from score
                    where batch = '2312'
                      and alert != 'Pas d''alerte');

delete
from entreprise_pge
where siren not in (select siren
                    from score
                    where batch = '2312'
                      and alert != 'Pas d''alerte');

delete
from etablissement_procol
where siren not in (select siren
                    from score
                    where batch = '2312'
                      and alert != 'Pas d''alerte');

delete from etablissement_comments;
delete from etablissement_follow;
delete
from entreprise_ellisphere
where siren not in (select siren
                    from score
                    where batch = '2312'
                      and alert != 'Pas d''alerte');

delete
from entreprise_diane
where siren not in (select siren
                    from score
                    where batch = '2312'
                      and alert != 'Pas d''alerte');

select count(*) from v_summaries