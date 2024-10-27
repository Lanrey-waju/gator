-- name: CreateFeedFollow :one
WITH new_follow AS (
INSERT INTO feed_follows
    (
    id, created_at, updated_at, user_id, feed_id
    )
VALUES
    (
        $1, $2, $3, $4, $5
    )
RETURNING *
)
SELECT nf.id, nf.created_at, nf.updated_at, nf.user_id, nf.feed_id, u.name AS follower, f.name as following
FROM new_follow nf JOIN users u ON nf.user_id = u.id JOIN feeds f ON nf.feed_id = f.id;

-- name: GetFeedFollowsForUser :many
SELECT ff.id, u.name AS follower, f.name AS feed
FROM feed_follows ff
    JOIN users u ON ff.user_id = u.id
    JOIN feeds f ON ff.feed_id = f.id
WHERE ff.user_id = $1;


-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE feed_follows.user_id = $1 AND
    feed_follows.feed_id = $2;