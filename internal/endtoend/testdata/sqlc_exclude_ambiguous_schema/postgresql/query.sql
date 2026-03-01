-- Partially qualified exclude "posts.user_id" is ambiguous when both
-- public.posts and enterprise.posts have a user_id column
-- name: CrossSchemaExcludeAmbiguous :one
SELECT public.posts.*, enterprise.posts.*, sqlc.exclude('posts.user_id')
FROM public.posts
JOIN enterprise.posts ON public.posts.id = enterprise.posts.id
WHERE public.posts.id = $1;
