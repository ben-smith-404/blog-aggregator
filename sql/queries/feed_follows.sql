-- name: CreateFeedFollower :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
        VALUES (
            $1,
            $2,
            $3,
            $4,
            $5
        )
        RETURNING *
    )
SELECT inserted_feed_follow.*, u.name as user_name, f.name as feed_name
FROM inserted_feed_follow
INNER JOIN users u on user_id = u.id
INNER JOIN feeds f ON feed_id = f.id;

-- name: GetFeedsUserFollows :many
SELECT ff.*, u.name as user_name, f.name as feed_name 
FROM feed_follows ff 
INNER JOIN users u ON ff.user_id = u.id
INNER JOIN feeds f ON ff.feed_id = f.id
WHERE ff.user_id = $1;

-- name: DeleteFollowedFeed :exec
DELETE FROM feed_follows 
WHERE user_id = $1
AND feed_id = $2;