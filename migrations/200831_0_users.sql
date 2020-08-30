create table users (
  id varchar(36) primary key,
  userName text,
  firstName text,
  lastName text,
  roles text[]
);

create table roles (
  code text,
  description text
);