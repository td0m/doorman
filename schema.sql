create table tuples(
  object text not null,
  relation text not null,
  user_id text not null,

  primary key(object, relation, user_id)
);

