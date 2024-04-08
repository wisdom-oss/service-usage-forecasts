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
       date,
       usage_type,
       amount
FROM wisdom.water_usage.usages
WHERE municipality ~ $1;

-- name: get-usages-by-municipality-consumer-groups
SELECT municipality,
       date,
       usage_type,
       amount
FROM wisdom.water_usage.usages
WHERE municipality ~ $1
  AND usage_type IN ($2);