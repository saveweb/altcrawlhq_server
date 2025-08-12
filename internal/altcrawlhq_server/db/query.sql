-- -- name: GetURL :one
-- SELECT * FROM urls
-- WHERE project = ? AND id = ? LIMIT 1;

-- index: urls_project_status
-- name: GetFreshURLs :many
SELECT * FROM urls
WHERE project = ? AND status = 'FRESH'
LIMIT ?;

-- index: PRIMARY KEY
-- name: ClaimThisURL :exec
UPDATE urls
SET status = 'CLAIMED', timestamp = strftime('%s', 'now')
WHERE project = ? AND id = ?
RETURNING *;

-- name: CreateURL :exec
INSERT INTO urls (project, id, value, via, host, path, type, crawler, status, lift_off, timestamp)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- index: PRIMARY KEY
-- name: DoneURL :exec
UPDATE urls
SET status = 'DONE', timestamp = strftime('%s', 'now')
WHERE project = ? AND id = ?;

-- index: PRIMARY KEY
-- name: DeleteURL :exec
DELETE FROM urls
WHERE project = ? AND id = ?;


-- index: PRIMARY KEY
-- name: ResetURL :exec
UPDATE urls
SET status = 'FRESH', timestamp = strftime('%s', 'now')
WHERE project = ? AND id = ?;

-- index: PRIMARY KEY
-- name: CountSeen :one
SELECT COUNT(*) FROM seens
WHERE project = ? AND type = ? AND value = ?;

-- name: CreateSeen :exec
INSERT INTO seens (project, type, value)
VALUES (?, ?, ?);

-- name: RefreshSeen :exec
UPDATE seens
SET timestamp = strftime('%s', 'now')
WHERE project = ? AND type = ? AND value = ?;