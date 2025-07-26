-- name: GetEvents :many
select *
from events
;

-- name: GetNextSeasonStart :one
select start_time
from seasons
where start_time > ?
order by start_time asc
limit 1;

-- name: GetSeasonStart :one
select start_time
from seasons
order by start_time asc
limit 1 offset ?;

-- name: GetSeasonID :one
select count(*) as season_id
from seasons
where start_time <= ?;

-- name: GetUserSwinces :many
select s.*, e.time
from swinces s
join events e on s.event_id = e.event_id
where s.participant_id = ? and e.time >= ? and e.time < ?;

-- name: GetUserNominations :many
select s.*, e.time
from swinces s
join events e on s.event_id = e.event_id
where s.participant_id = ? and s.nominee_id is not null and s.fulfillment_id is not null
  and e.time >= ? and e.time < ?;

-- name: GetUserFulfillments :many
select s.*, e.time
from swinces s
join events e on s.event_id = e.event_id
where s.participant_id = ? and s.swince_id in (
  select fulfillment_id from swinces where fulfillment_id is not null
) and e.time >= ? and e.time < ?;

-- name: GetAllUsersInSeason :many
select distinct participant_id
from swinces s
join events e on s.event_id = e.event_id
where e.time >= ? and e.time < ?;

