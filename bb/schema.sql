BEGIN;

CREATE TABLE IF NOT EXISTS accounts (
    account_id  INTEGER PRIMARY KEY,
    login       TEXT NOT NULL UNIQUE,
	provider    TEXT NOT NULL, -- oauth2 provider name
    created     TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    key             TEXT PRIMARY KEY,
    account         INTEGER REFERENCES accounts(account_id),
    created         TIMESTAMPTZ NOT NULL,
	provider        TEXT NOT NULL,
	scopes          TEXT NOT NULL,
    access_token    TEXT NOT NULL
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
