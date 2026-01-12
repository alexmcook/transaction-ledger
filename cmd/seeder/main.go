package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func makePools(numShards int) ([]*pgxpool.Pool, error) {
	pools := make([]*pgxpool.Pool, numShards)

	for i := range numShards {
		seedUrl, ok := os.LookupEnv("SEED_URL")
		if !ok {
			return nil, fmt.Errorf("Database URL not set")
		}

		dbUrl := fmt.Sprintf(seedUrl, "localhost", 5432+i)

		pool, err := pgxpool.New(context.Background(), dbUrl)
		if err != nil {
			return nil, fmt.Errorf("Failed to connect to database shard: %v", err)
		}

		pools[i] = pool

		if err := pool.Ping(context.Background()); err != nil {
			return nil, fmt.Errorf("Failed to ping database: %v", err)
		}
	}

	return pools, nil
}

func makeUUIDs(n int) ([]uuid.UUID, error) {
	uuids := make([]uuid.UUID, 0, n)
	for range n {
		uid, err := uuid.NewV7()
		if err != nil {
			return nil, fmt.Errorf("Failed to generate UUIDv7: %v", err)
		}
		uuids = append(uuids, uid)
	}
	return uuids, nil
}

func getShard(uid uuid.UUID, numShards int) int {
	val := binary.BigEndian.Uint64(uid[8:16])
	return int(val % uint64(numShards))
}

func truncateTables(pools []*pgxpool.Pool) error {
	for _, pool := range pools {
		_, err := pool.Exec(context.Background(), `TRUNCATE TABLE accounts`)
		if err != nil {
			return fmt.Errorf("Failed to truncate accounts table: %v", err)
		}
	}
	return nil
}

func seedDatabase(pools []*pgxpool.Pool, uuids []uuid.UUID) error {
	numShards := len(pools)
	for _, uid := range uuids {
		shardIndex := getShard(uid, numShards)
		pool := pools[shardIndex]

		_, err := pool.Exec(context.Background(),
			`INSERT INTO accounts (id, balance, created_at) VALUES ($1, $2, NOW())`,
			uid, 1000)
		if err != nil {
			return fmt.Errorf("Failed to insert account: %v", err)
		}
	}
	return nil
}

func writeToFile(filename string, data []uuid.UUID) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Failed to create file: %v", err)
	}
	defer file.Close()

	for _, uid := range data {
		_, err := file.WriteString(uid.String() + "\n")
		if err != nil {
			return fmt.Errorf("Failed to write to file: %v", err)
		}
	}
	return nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading .env file: %v\n", err)
		os.Exit(1)
	}

	numShards := 2
	pools, err := makePools(numShards)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up pools: %v\n", err)
		os.Exit(1)
	}

	uuids, err := makeUUIDs(10000)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating UUIDs: %v\n", err)
		os.Exit(1)
	}

	if err = truncateTables(pools); err != nil {
		fmt.Fprintf(os.Stderr, "Error truncating tables: %v\n", err)
		os.Exit(1)
	}

	if err = seedDatabase(pools, uuids); err != nil {
		fmt.Fprintf(os.Stderr, "Error seeding data: %v\n", err)
		os.Exit(1)
	}

	if err = writeToFile("data/account_ids.csv", uuids); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing UUIDs to file: %v\n", err)
		os.Exit(1)
	}

	for _, pool := range pools {
		pool.Close()
	}

	fmt.Println("Database seeding completed successfully")
}
