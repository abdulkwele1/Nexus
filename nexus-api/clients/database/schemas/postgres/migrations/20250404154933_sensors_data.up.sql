CREATE TABLE sensors (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    location VARCHAR(255) NOT NULL,
    installation_date TIMESTAMP WITH TIME ZONE NOT NULL
);
