CREATE TABLE IF NOT EXISTS login_authentications
(
    id               BIGSERIAL,
    user_name        TEXT UNIQUE NOT NULL,
    password_hash    TEXT        NOT NULL
);
