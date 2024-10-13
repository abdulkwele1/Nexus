CREATE TABLE IF NOT EXISTS login_cookies
(
    cookie           TEXT UNIQUE NOT NULL,
    user_name        TEXT NOT NULL,
    expires_at       TIMESTAMPTZ NOT NULL
);
