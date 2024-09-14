create table code_group
(
    id         serial primary key,
    owner_id   text      not null,
    group_id   text      not null,
    name       text      not null,
    created_at timestamp not null default now(),
    constraint owner_id_group_id_unique
        unique (owner_id, group_id),
    constraint owner_id_name_unique
        unique (owner_id, name)
);

create table code
(
    id            serial primary key,
    code_group_id integer not null
    -- TODO
);
