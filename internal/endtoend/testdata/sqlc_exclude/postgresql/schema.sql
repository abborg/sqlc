CREATE TABLE users (
    id integer NOT NULL PRIMARY KEY,
    name varchar(255) NOT NULL,
    email varchar(255) NOT NULL,
    password varchar(255) NOT NULL
);

CREATE TABLE posts (
    id integer NOT NULL PRIMARY KEY,
    user_id integer NOT NULL REFERENCES users(id),
    title varchar(255) NOT NULL,
    api_key varchar(255) NOT NULL
);
