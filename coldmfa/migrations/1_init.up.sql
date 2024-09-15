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
    id             serial primary key,
    code_group_id  integer   not null,
    code_id        text      not null,

    original       text      not null,   -- The original information provided to seed the code

    name           text      not null,
    preferred_name text,                 -- The name chosen by the user

    created_at     timestamp not null default now(),

    constraint code_group_id_code_id_unique
        unique (code_group_id, code_id), -- Note only unique within the group, so querying by code_id must only be done in the context of a group
    constraint code_group_id_original_unique
        unique (code_group_id, original),
    constraint code_group_id_preferred_name_unique
        unique (code_group_id, preferred_name)
);
