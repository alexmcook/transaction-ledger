CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  created_at BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS accounts (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  balance BIGINT NOT NULL,
  created_at BIGINT NOT NULL
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
  created_at BIGINT NOT NULL,
  bucket_id SMALLINT NOT NULL
) PARTITION BY LIST (bucket_id);

-- Create two partitions for write-behind worker to efficiently prune transactions
CREATE TABLE IF NOT EXISTS tx_buf_0 PARTITION OF transactions FOR VALUES IN (0);
CREATE TABLE IF NOT EXISTS tx_buf_1 PARTITION OF transactions FOR VALUES IN (1);

-- Default partition to catch bad bucket_id values
CREATE TABLE IF NOT EXISTS tx_default PARTITION OF transactions DEFAULT;

-- Index on account_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);
