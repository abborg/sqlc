-- Unqualified exclude: user_id appears in both posts and comments - ambiguous
-- name: JoinExcludeAmbiguous :one
SELECT users.*, posts.*, comments.*, sqlc.exclude('user_id')
FROM users
JOIN posts ON users.id = posts.user_id
JOIN comments ON posts.id = comments.post_id
WHERE comments.id = $1;

-- Partially qualified exclude: posts.user_id matches both public.posts and enterprise.posts - ambiguous
-- name: CrossSchemaExcludeAmbiguous :one
SELECT public.posts.*, enterprise.posts.*, sqlc.exclude('posts.user_id')
FROM public.posts
JOIN enterprise.posts ON public.posts.id = enterprise.posts.id
WHERE public.posts.id = $1;
