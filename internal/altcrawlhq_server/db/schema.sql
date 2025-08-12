CREATE TABLE IF NOT EXISTS urls (
    project TEXT NOT NULL,
    id TEXT NOT NULL,
    value TEXT NOT NULL,
    via TEXT DEFAULT '' NOT NULL,
    host TEXT DEFAULT '' NOT NULL,
    path TEXT DEFAULT '' NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('seed', 'asset')),
    crawler TEXT DEFAULT '' NOT NULL,
    status TEXT NOT NULL DEFAULT 'FRESH' CHECK (status IN ('FRESH', 'CLAIMED', 'DONE')),
    timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    PRIMARY KEY (project, id)
);
CREATE UNIQUE INDEX IF NOT EXISTS urls_project_value ON urls (project, type, value); -- for deduplication
CREATE INDEX IF NOT EXISTS urls_project_status ON urls (project, status); -- for queue


-- seens is for assets only
CREATE TABLE IF NOT EXISTS seens (
    project TEXT NOT NULL,
    type TEXT NOT NULL,
    value TEXT NOT NULL,
    timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    PRIMARY KEY (project, type, value)
);
