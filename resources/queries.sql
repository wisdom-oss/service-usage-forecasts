-- name: get-consumer-groups
SELECT id,
    name,
    description,
    external_identifier
FROM wisdom.water_usage.usage_types;

-- name: get-consumer-groups-by-external-id
SELECT id,
    name,
    description,
    external_identifier
FROM wisdom.water_usage.usage_types
WHERE external_identifier = ANY ($1);

-- name: get-usages-by-municipality
SELECT municipality,
    time,
    usage_type,
    amount
FROM wisdom.timeseries.water_usage
WHERE municipality ~ $1;

-- name: get-usages-by-municipality-consumer-groups
SELECT municipality,
    time,
    usage_type,
    amount
FROM wisdom.timeseries.water_usage
WHERE municipality ~ $1
  AND usage_type IN ($2);

-- name: get-bucketed-usages-by-municipality
SELECT municipality,
    time_bucket($1, time) AS time,
    usage_type,
    SUM(amount)           AS amount
FROM wisdom.timeseries.water_usage
WHERE municipality ~ $2
GROUP BY time, municipality, usage_type
ORDER BY time;

-- name: get-bucketed-usages-by-municipality-consumer-groups
SELECT municipality,
    time_bucket($1, time) AS time,
    usage_type,
    SUM(amount)           AS amount
FROM wisdom.timeseries.water_usage
WHERE municipality ~ $2
  AND usage_type IN ($3)
GROUP BY time, municipality, usage_type
ORDER BY time;