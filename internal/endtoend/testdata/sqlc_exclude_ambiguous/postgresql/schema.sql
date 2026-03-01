CREATE TABLE users (
    id integer NOT NULL PRIMARY KEY,
    name varchar(255) NOT NULL,
    email varchar(255) NOT NULL
);

CREATE TABLE posts (
    id integer NOT NULL PRIMARY KEY,
    user_id integer NOT NULL REFERENCES users(id),
    title varchar(255) NOT NULL
);

CREATE TABLE comments (
    id integer NOT NULL PRIMARY KEY,
    post_id integer NOT NULL REFERENCES posts(id),
    user_id integer NOT NULL REFERENCES users(id),
    body text NOT NULL
);

-- Cross-schema ambiguous exclude: posts.user_id matches both public.posts and enterprise.posts
CREATE SCHEMA enterprise;

CREATE TABLE enterprise.posts (
    id integer NOT NULL PRIMARY KEY,
    user_id integer NOT NULL,
    title varchar(255) NOT NULL
);
