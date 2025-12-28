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
INSERT INTO transaction_types (type) VALUES ('credit'), ('debit');

CREATE TABLE IF NOT EXISTS transactions (
  id UUID PRIMARY KEY,
  account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  amount BIGINT NOT NULL,
  transaction_type SMALLINT NOT NULL REFERENCES transaction_types(id),
  created_at BIGINT NOT NULL
);
