BEGIN;

CREATE TYPE account_role AS ENUM ('admin', 'service', 'user');


CREATE TABLE IF NOT EXISTS accounts (
    id              SERIAL PRIMARY KEY,
    login           TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    role            account_role NOT NULL,
    valid_till      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL
);


COMMIT;
