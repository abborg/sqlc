CREATE SCHEMA enterprise;

CREATE TABLE public.posts (
    id integer NOT NULL PRIMARY KEY,
    user_id integer NOT NULL,
    title varchar(255) NOT NULL
);

CREATE TABLE enterprise.posts (
    id integer NOT NULL PRIMARY KEY,
    user_id integer NOT NULL,
    title varchar(255) NOT NULL
);
