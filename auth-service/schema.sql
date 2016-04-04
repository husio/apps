BEGIN;


CREATE TABLE IF NOT EXISTS accounts (
    id              SERIAL PRIMARY KEY,
    login           TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    scopes          TEXT[],
    valid_till      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL
);


COMMIT;
