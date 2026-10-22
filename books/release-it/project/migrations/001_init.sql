-- Registry schema: tenants, keys, nodes, databases, placement, usage ledger.

CREATE TABLE tenants (
    id         text PRIMARY KEY,
    name       text NOT NULL,
    plan       text NOT NULL DEFAULT 'free',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE api_keys (
    key_hash  text PRIMARY KEY,
    tenant_id text NOT NULL REFERENCES tenants(id),
    admin     boolean NOT NULL DEFAULT false
);

CREATE TABLE nodes (
    name              text PRIMARY KEY,
    url               text NOT NULL,
    last_heartbeat_at timestamptz,
    fds               int NOT NULL DEFAULT 0,
    disk_free_bytes   bigint NOT NULL DEFAULT 0
);

CREATE TABLE databases (
    id          uuid PRIMARY KEY,
    tenant_id   text NOT NULL REFERENCES tenants(id),
    name        text NOT NULL,
    state       text NOT NULL DEFAULT 'provisioning',
    bytes_total int NOT NULL DEFAULT 0,
    rows_total  int NOT NULL DEFAULT 0,
    size_bytes  bigint NOT NULL DEFAULT 0,
    last_backup_at timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, name)
);

CREATE TABLE placement (
    database_id uuid PRIMARY KEY REFERENCES databases(id),
    owner_node  text NOT NULL REFERENCES nodes(name),
    epoch       int NOT NULL DEFAULT 1
);

CREATE TABLE usage_ledger (
    id          bigserial PRIMARY KEY,
    at          timestamptz NOT NULL DEFAULT now(),
    tenant_id   text NOT NULL,
    database_id uuid NOT NULL,
    kind        text NOT NULL,
    statements  int NOT NULL DEFAULT 0,
    rows        int NOT NULL DEFAULT 0,
    bytes       int NOT NULL DEFAULT 0
);
