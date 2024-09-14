create table code_group
(
    id         serial primary key,
    group_id   text      not null,
    name       text      not null,
    created_at timestamp not null default now()
);

create table code
(
    id            serial primary key,
    code_group_id integer not null
    -- TODO
);
