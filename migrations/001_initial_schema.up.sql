CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS accounts (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  balance BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS transaction_types (
  id SMALLSERIAL PRIMARY KEY,
  type VARCHAR(12) NOT NULL UNIQUE
);

-- Prepopulate transaction types, 'credit' 1 and 'debit' 2
INSERT INTO transaction_types (type) VALUES ('credit'), ('debit');

-- Omit PRIMARY KEY for performance, UUID likelihood of collision is effectively zero
CREATE TABLE IF NOT EXISTS transactions (
  id UUID NOT NULL,
  account_id UUID NOT NULL,
  amount BIGINT NOT NULL,
  transaction_type SMALLINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  partition_key SMALLINT NOT NULL
) PARTITION BY LIST (partition_key);

-- Create two partitions for write-behind worker to efficiently prune transactions
CREATE TABLE IF NOT EXISTS transactions_p0 PARTITION OF transactions FOR VALUES IN (0);
CREATE TABLE IF NOT EXISTS transactions_p1 PARTITION OF transactions FOR VALUES IN (1);

-- Index on account_id for faster lookups by worker
CREATE INDEX IF NOT EXISTS idx_transactions_p0_worker ON transactions_p0 (account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_p1_worker ON transactions_p1 (account_id);
