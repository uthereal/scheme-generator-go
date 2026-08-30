CREATE TABLE auth.devices (
	id UUID PRIMARY KEY,
	user_id INT REFERENCES auth.users(id),
	session_token UUID
);
