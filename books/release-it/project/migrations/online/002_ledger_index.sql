-- Covering index for the usage report: the 24 h window becomes an index-only
-- scan instead of a pass over every ledger row. Built online, without
-- blocking writes, by `make migrate` before a rollout; never at startup
-- (CREATE INDEX CONCURRENTLY cannot run inside the startup migrator's
-- transaction, and an index build does not belong in a deploy's critical path).
CREATE INDEX CONCURRENTLY IF NOT EXISTS usage_ledger_at_covering
    ON usage_ledger (at) INCLUDE (tenant_id, statements, rows, bytes);
