BEGIN;


CREATE TABLE IF NOT EXISTS accounts (
    account_id      SERIAL PRIMARY KEY,
    name            TEXT NOT NULL,
    created 		TIMESTAMPTZ NOT NULL,
    updated         TIMESTAMPTZ NOT NULL
);


CREATE TABLE IF NOT EXISTS topics (
    topic_id 	    SERIAL PRIMARY KEY,
    author_id       INTEGER NOT NULL REFERENCES accounts(account_id),
    title 			TEXT NOT NULL,
    tags            TEXT[] NOT NULL,
    created 		TIMESTAMPTZ NOT NULL,
    updated         TIMESTAMPTZ NOT NULL
);


CREATE TABLE IF NOT EXISTS comments (
    comment_id 	    SERIAL PRIMARY KEY,
    topic_id        INTEGER NOT NULL REFERENCES topics(topic_id),
    author_id       INTEGER NOT NULL REFERENCES accounts(account_id),
    content         TEXT NOT NULL,
    created 		TIMESTAMPTZ NOT NULL,
    updated         TIMESTAMPTZ NOT NULL
);


COMMIT;
