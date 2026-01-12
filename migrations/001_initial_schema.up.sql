CREATE TABLE IF NOT EXISTS accounts (
  id UUID PRIMARY KEY,
  balance BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions_history (
  id UUID PRIMARY KEY,
  account_id UUID NOT NULL,
  amount BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

DO $$
BEGIN
  FOR i IN 0..63 LOOP
    EXECUTE format('
      CREATE TABLE IF NOT EXISTS transactions_%s (
        id UUID PRIMARY KEY,
        account_id UUID NOT NULL,
        amount BIGINT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL
      );

      CREATE UNLOGGED TABLE IF NOT EXISTS staging_%s (
        id UUID,
        account_id UUID,
        amount BIGINT,
        created_at TIMESTAMPTZ
      );

      ALTER TABLE staging_%s SET (autovacuum_enabled = false);
    ', i, i, i, i, i);
  END LOOP;
END $$;

-- Table to manually track Kafka offsets for each partition
CREATE TABLE IF NOT EXISTS kafka_offsets (
  partition_id INT PRIMARY KEY,
  last_offset BIGINT NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Initialize kafka_offsets for partitions 0 to 63 with last_offset -1
-- This assumes a Kafka topic with 64 partitions
INSERT INTO kafka_offsets (partition_id, last_offset)
SELECT p_id, -1 FROM generate_series(0, 63) AS p_id
ON CONFLICT (partition_id) DO NOTHING;
