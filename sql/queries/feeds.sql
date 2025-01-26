-- name: AddFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: FetchFeeds :many
SELECT
    feeds.id AS feed_id,
    feeds.name AS feed_name,
    feeds.url AS feed_url,
    users.name AS user_name
FROM feeds
JOIN users
ON users.id = feeds.user_id;

-- name: CreateFeedFollow :many
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES (gen_random_uuid(), NOW(), NOW(), $1, $2)
    RETURNING *
)

SELECT
    inserted_feed_follow.*,
    feeds.name AS feed_name,
    users.name AS user_name
FROM inserted_feed_follow
INNER JOIN feeds
ON inserted_feed_follow.feed_id = feeds.id
INNER JOIN users
ON inserted_feed_follow.user_id = users.id;

-- name: GetFeedFollow :one
SELECT *
FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;

-- name: GetFeedFollowsForUser :many
SELECT 
    feed_follows.*,
    feeds.name AS feed_name,
    users.name AS user_name
FROM feed_follows
INNER JOIN feeds
ON feed_follows.feed_id = feeds.id
INNER JOIN users
ON feed_follows.user_id = users.id
WHERE feed_follows.user_id = $1;

-- name: Unfollow :exec
DELETE FROM feed_follows
    USING feeds
    WHERE feed_follows.feed_id = feeds.id
    AND feed_follows.user_id = $1
    AND feeds.name = $2
    RETURNING feed_follows.*;

-- name: MarkFeedFetched :one
UPDATE feeds
SET last_fetched_at = NOW(),
updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetNextFeedToFetch :one
SELECT *
FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;
