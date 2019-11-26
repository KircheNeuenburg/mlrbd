create table schema_version (
    version text not null
);

create table room_group_map ( 
    id integer not null generated always as identity,
    ldap_group text unique,
    matrix_room text unique
);

