create sequence logs_id;
create table logs (
  id integer primary key default nextval('logs_id'),
  date_add timestamp default current_timestamp,
  path text,
  method text,
  body text,
  token text
);