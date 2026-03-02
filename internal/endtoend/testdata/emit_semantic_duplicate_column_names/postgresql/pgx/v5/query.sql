-- name: RightJoin :many
SELECT f.id, f.bar_id, b.id
FROM foo f
RIGHT JOIN bar b ON b.id = f.bar_id
WHERE f.id = $1;

-- name: TableName :one
SELECT foo.id
FROM foo
JOIN bar ON foo.bar_id = bar.id
WHERE bar.id = $1 AND foo.id = $2;

-- name: AliasJoin :many
SELECT p.user_id, c.user_id
FROM posts p
JOIN comments c ON c.post_id = p.id;

-- name: AliasJoinLong :many
SELECT pt.user_id, ct.user_id
FROM posts pt
JOIN comments ct ON ct.post_id = pt.id;
