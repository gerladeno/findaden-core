-- noinspection SqlNoDataSourceInspectionForFile


-- +migrate Up

create table config
(
    uuid text not null
        primary key,
    created timestamp default now(),
    updated timestamp default now()
);

create table settings
(
    uuid  text not null
        primary key
        constraint fk_configs_settings
            references config,
    theme int
);

create table personal
(
    uuid        text not null
        primary key
        constraint fk_configs_personal
            references config,
    username    text,
    avatar_link text,
    gender      smallint,
    age         smallint
);

create table regions
(
    id          bigserial
        primary key,
    created_at  timestamp with time zone,
    updated_at  timestamp with time zone,
    deleted_at  timestamp with time zone,
    name        text,
    description text
);

create table relations
(
    uuid     text not null
        constraint fk_configs_relations_uuid
            references config,
    target   text not null
        constraint fk_configs_relations_target
            references config,
    relation smallint,
    primary key (uuid, target)
);

create table search_criteria
(
    uuid       text not null
        primary key
        constraint fk_configs_criteria
            references config,
    price_from numeric,
    price_to   numeric,
    gender     smallint,
    age_from   numeric,
    age_to     numeric
);



create table uuid_regions
(
    uuid text   not null
        constraint fk_uuid_regions_search_criteria
            references search_criteria,
    region_id            bigint not null
        constraint fk_uuid_regions_region
            references regions,
    primary key (uuid, region_id)
);

-- +migrate Down

DROP TABLE config CASCADE;
DROP TABLE settings CASCADE;
DROP TABLE personal CASCADE;
DROP TABLE regions CASCADE;
DROP TABLE relations CASCADE;
DROP TABLE search_criteria CASCADE;
DROP TABLE uuid_regions CASCADE;