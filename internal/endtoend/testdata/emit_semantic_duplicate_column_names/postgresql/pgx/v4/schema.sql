CREATE TABLE bar (id serial not null unique);
CREATE TABLE foo (id serial not null, bar_id int references bar(id));
CREATE TABLE posts (id serial primary key, user_id int not null);
CREATE TABLE comments (id serial primary key, user_id int not null, post_id int references posts(id));
