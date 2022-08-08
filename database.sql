create table hearings
(
    id         INTEGER primary key,
    link       TEXT    default '' not null unique,
    topics     TEXT    default '',
    proposals  TEXT    default '',
    place      TEXT    default '',
    date       TEXT    default '1970-01-01 00:00:00',
    published  BOOLEAN default false,
    raw        TEXT    default '',
    created_at TEXT    default '1970-01-01 00:00:00'
);

