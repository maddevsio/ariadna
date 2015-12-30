create table road (
    id serial not null primary key,
    node_id bigint not null,
    country varchar(100) null,
    name varchar(200) null,
    city varchar(100) null,
    housenumber varchar(128) null,
    street varchar(255) null,
    coords geometry,
    centroid geometry
);
