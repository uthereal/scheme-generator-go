CREATE TABLE content.tags (
	id BIGSERIAL PRIMARY KEY,
	slug VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE content.posts (
	id BIGSERIAL PRIMARY KEY,
	author_id BIGINT NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
	title VARCHAR(255) NOT NULL,
	content TEXT NOT NULL,
	rating NUMERIC(5, 2) NOT NULL DEFAULT 0.0,
	published_at TIMESTAMPTZ
);

ALTER TABLE content.posts RENAME COLUMN author_id TO user_id;

CREATE TABLE content.comments (
	id BIGSERIAL PRIMARY KEY,
	post_id BIGINT NOT NULL REFERENCES content.posts(id) ON DELETE CASCADE,
	user_id BIGINT NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
	text TEXT NOT NULL
);

CREATE TABLE content.post_tags (
	post_id BIGINT NOT NULL REFERENCES content.posts(id) ON DELETE CASCADE,
	tag_id BIGINT NOT NULL REFERENCES content.tags(id) ON DELETE CASCADE,
	PRIMARY KEY (post_id, tag_id)
);
