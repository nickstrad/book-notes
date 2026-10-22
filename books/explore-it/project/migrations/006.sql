-- A cancellation observation authenticates the revoked token without retaining
-- an active lease. Legacy cancelled jobs have no observation token.
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS cancelled_lease_hash text;
