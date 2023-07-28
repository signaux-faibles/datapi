select date_add,
       path,
       method,
--        body,
       tokencontent ->> 'preferred_username'::text as username,
       COALESCE( tokencontent ->> 'segment'::text, '-' )  as segment,
       translate(((tokencontent ->> 'resource_access')::json ->> 'signauxfaibles')::json ->> 'roles', '[]', '{}')::text[]  as roles
from v_log
where date_add >= $1 and date_add < $2
order by date_add desc;
