-- name: GetUserWithoutPassword :one
SELECT *, sqlc.exclude('password') FROM users WHERE id = $1;

-- name: GetUserExcludeMultiple :one
SELECT *, sqlc.exclude('password'), sqlc.exclude('email') FROM users WHERE id = $1;

-- name: GetUserExcludeWithEmbed :one
SELECT sqlc.embed(users), sqlc.exclude('password') FROM users WHERE id = $1;

-- name: JoinEmbedExclude :one
SELECT sqlc.embed(users), sqlc.embed(posts), sqlc.exclude('password'), sqlc.exclude('api_key')
FROM posts
INNER JOIN users ON posts.user_id = users.id
WHERE posts.id = $1;

-- name: JoinEmbedExcludeQualified :one
SELECT sqlc.embed(users), sqlc.embed(posts), sqlc.exclude('users.password'), sqlc.exclude('posts.api_key')
FROM posts
INNER JOIN users ON posts.user_id = users.id
WHERE posts.id = $1;

-- name: JoinEmbedExcludeAlias :one
-- Uses table aliases (u, p) but exclude with table names - validates join+embed+exclude with aliased from clause
SELECT sqlc.embed(u), sqlc.embed(p), sqlc.exclude('users.password'), sqlc.exclude('posts.api_key')
FROM posts p
INNER JOIN users u ON p.user_id = u.id
WHERE p.id = $1;
