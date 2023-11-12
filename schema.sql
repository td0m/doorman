create table tuples(
  object text not null,
  relation text not null,
  subject text not null,

  primary key(object, relation, subject)
);

