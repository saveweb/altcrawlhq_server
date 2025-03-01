-- name: GetURL :one
SELECT * FROM urls
WHERE project = ? AND id = ? LIMIT 1;

-- name: GetFreshURLs :many
SELECT * FROM urls
WHERE project = ? AND status = 'FRESH'
ORDER BY lift_off ASC
LIMIT ?;

-- name: ClaimThisURL :exec
UPDATE urls
SET status = 'CLAIMED', timestamp = strftime('%s', 'now')
WHERE project = ? AND id = ?
RETURNING *;

-- name: CreateURL :one
INSERT INTO urls (project, id, value, via, host, path, type, crawler, status, lift_off, timestamp)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: DoneURL :exec
UPDATE urls
SET status = 'DONE', timestamp = strftime('%s', 'now')
WHERE project = ? AND id = ?;


-- name: ResetURL :exec
UPDATE urls
SET status = 'FRESH', timestamp = strftime('%s', 'now')
WHERE project = ? AND id = ?;

-- name: CountSeen :one
SELECT COUNT(*) FROM seens
WHERE project = ? AND type = ? AND value = ?;

-- name: CreateSeen :exec
INSERT INTO seens (project, type, value)
VALUES (?, ?, ?);