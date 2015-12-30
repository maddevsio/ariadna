create table road_intersection (
    id serial not null primary key,
    inter_count bigint not null,
    name varchar(200) null,
    coords geometry
);


INSERT INTO road_intersection (coords, inter_count, name)
(SELECT
    ST_Intersection(a.coords, b.coords),
    Count(Distinct a.node_id),
    concat(a.name, ' ', b.name) as InterName

FROM
    road as a,
    road as b
WHERE
    ST_Touches(a.coords, b.coords)
    AND a.node_id != b.node_id
GROUP BY
    ST_Intersection(a.coords, b.coords),
    InterName
);