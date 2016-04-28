CREATE TABLE IF NOT EXISTS bookmarks (
	bookmark_id     SERIAL PRIMARY KEY,
	title           TEXT NOT NULL,
	url 			TEXT NOT NULL,
	created         TIMESTAMPTZ NOT NULL
);
