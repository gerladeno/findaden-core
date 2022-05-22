-- noinspection SqlNoDataSourceInspectionForFile


-- +migrate Up


create table message
(
    sender    text not null
        constraint fk_sender_uuid
            references config,
    receiver  text not null
        constraint fk_receiver_uuid
            references config,
    timestamp timestamp default now(),
    body      text
);

create index message_idx
    on message (sender, receiver);

create table chat
(
    uuid1 text not null
        constraint fk_uuid1
            references config,
    uuid2 text not null
        constraint fk_uuid2
            references config,
    primary key (uuid1, uuid2),
    created timestamp default now(),
    updated timestamp default now()
);

create index uuid1_idx on chat(uuid1);
create index uuid2_idx on chat(uuid2);

-- +migrate Down

DROP TABLE message CASCADE;
DROP TABLE chat CASCADE;
