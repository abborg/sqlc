-- name: GetUserPosts :many
-- @group-by users.id
SELECT sqlc.embed(users), sqlc.embed_many(posts) FROM users
LEFT JOIN posts ON posts.user_id = users.id
ORDER BY users.id;
