CREATE SCHEMA auth;
CREATE SCHEMA content;
CREATE SCHEMA tenant;

CREATE TYPE auth.user_status AS ENUM ('active', 'inactive', 'suspended');
CREATE CAST (varchar AS auth.user_status) WITH INOUT AS IMPLICIT;
CREATE CAST (text AS auth.user_status) WITH INOUT AS IMPLICIT;

CREATE TABLE auth.roles (
	id BIGSERIAL PRIMARY KEY,
	name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE auth.permissions (
	id BIGSERIAL PRIMARY KEY,
	code VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE auth.users (
	id BIGSERIAL PRIMARY KEY,
	email VARCHAR(255) NOT NULL UNIQUE,
	age INTEGER,
	tags TEXT[] NOT NULL DEFAULT '{}',
	preferences JSONB,
	metadata JSONB NOT NULL DEFAULT '{}',
	status auth.user_status NOT NULL DEFAULT 'active',
	role_id BIGINT REFERENCES auth.roles(id) ON DELETE SET NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE auth.profiles (
	id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL UNIQUE REFERENCES auth.users(id) ON DELETE CASCADE,
	bio TEXT,
	location POINT,
	active_duration INTERVAL NOT NULL DEFAULT '0 seconds',
	is_public BOOLEAN
);

CREATE TABLE auth.role_permissions (
	role_id BIGINT NOT NULL REFERENCES auth.roles(id) ON DELETE CASCADE,
	permission_id BIGINT NOT NULL REFERENCES auth.permissions(id) ON DELETE CASCADE,
	PRIMARY KEY (role_id, permission_id)
);
