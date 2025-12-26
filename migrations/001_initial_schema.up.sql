CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  created_at BIGINT DEFAULT (floor(extract(epoch FROM now() AT TIME ZONE 'UTC') * 1000)::BIGINT)
);

CREATE TABLE IF NOT EXISTS accounts (
  id BIGSERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id) ON DELETE CASCADE,
  balance BIGINT DEFAULT 0,
  created_at BIGINT DEFAULT (floor(extract(epoch FROM now() AT TIME ZONE 'UTC') * 1000)::BIGINT)
);

CREATE TABLE IF NOT EXISTS transaction_types (
  id SMALLSERIAL PRIMARY KEY,
  type VARCHAR(12) UNIQUE NOT NULL
);
INSERT INTO transaction_types (type) VALUES ('deposit'), ('withdrawal');

CREATE TABLE IF NOT EXISTS transactions (
  id BIGSERIAL PRIMARY KEY,
  account_id INT REFERENCES accounts(id) ON DELETE CASCADE,
  amount BIGINT NOT NULL,
  transaction_type SMALLINT NOT NULL REFERENCES transaction_types(id),
  created_at BIGINT DEFAULT (floor(extract(epoch FROM now() AT TIME ZONE 'UTC') * 1000)::BIGINT)
);
