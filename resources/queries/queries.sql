-- name: GetSwinces :many
select *
from swinces;

-- name: GetUnfulfilledNominations :many
select s.swince_id, s.discord_id, s.nominates
from swinceurs s
where s.discord_id = ? and s.nomination_fulfilled = 0;

-- name: CreateSwince :one
insert into swinces (media)
values (?)
returning swince_id, uploaded_on;

-- name: CreateSwinceur :exec
insert into swinceurs (swince_id, discord_id, late_swince_tax, nominates, nomination_fulfilled)
values (?, ?, ?, ?, ?);

-- name: FulfillNomination :exec
update swinceurs 
set nomination_fulfilled = 1
where swince_id = ? and discord_id = ?;

