-- Daily rollup of the usage ledger. `control purge` moves raw ledger rows
-- older than the retention window into it, so the ledger stays the size of
-- the window while the history stays queryable.
CREATE TABLE usage_daily (
    tenant_id  text NOT NULL,
    day        date NOT NULL,
    statements bigint NOT NULL DEFAULT 0,
    rows       bigint NOT NULL DEFAULT 0,
    bytes      bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (tenant_id, day)
);
