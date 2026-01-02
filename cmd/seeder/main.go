package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func exportIDsToCSV(ids []uuid.UUID, filename string) error {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, id := range ids {
		_, err := file.WriteString(id.String() + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

func seedTransactions(ctx context.Context, pool *pgxpool.Pool, accountIds []uuid.UUID, n int, now int64) error {
	batch := &pgx.Batch{}
	for range n {
		transactionId, err := uuid.NewV7()
		if err != nil {
			return err
		}
		accountId := accountIds[rand.Intn(len(accountIds))]
		amount := int64(rand.Intn(200) - 100)
		var transactionType int
		if amount >= 0 {
			transactionType = 1 // credit
		} else {
			transactionType = 2 // debit
		}
		batch.Queue("INSERT INTO transactions (id, account_id, amount, transaction_type, created_at) VALUES ($1, $2, $3, $4, $5)", transactionId, accountId, amount, transactionType, now)
	}
	br := pool.SendBatch(ctx, batch)
	err := br.Close()
	if err != nil {
		return err
	}
	return nil
}

func seedAccounts(ctx context.Context, pool *pgxpool.Pool, userIds []uuid.UUID, n int, now int64) ([]uuid.UUID, error) {
	accountIds := make([]uuid.UUID, n)
	batch := &pgx.Batch{}
	for i := range n {
		accountId, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}
		accountIds[i] = accountId
		userId := userIds[rand.Intn(len(userIds))]
		batch.Queue("INSERT INTO accounts (id, user_id, balance, created_at) VALUES ($1, $2, $3, $4)", accountId, userId, int64(1000), now)
	}
	br := pool.SendBatch(ctx, batch)
	err := br.Close()
	if err != nil {
		return nil, err
	}
	return accountIds, nil
}

func seedUsers(ctx context.Context, pool *pgxpool.Pool, n int, now int64) ([]uuid.UUID, error) {
	userIds := make([]uuid.UUID, n)
	batch := &pgx.Batch{}
	for i := range n {
		userId, err := uuid.NewV7()
		if err != nil {
			return nil, err
		}
		userIds[i] = userId
		batch.Queue("INSERT INTO users (id, created_at) VALUES ($1, $2)", userId, now)
	}
	br := pool.SendBatch(ctx, batch)
	err := br.Close()
	if err != nil {
		return nil, err
	}
	return userIds, nil
}

func cleanDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, "TRUNCATE TABLE transactions, accounts, users RESTART IDENTITY CASCADE")
	return err
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	check(err)

	dbUrl := os.Getenv("DATABASE_URL")
	pool, err := pgxpool.New(ctx, dbUrl)
	check(err)
	defer pool.Close()

	fmt.Println("Connected to database successfully:", pool.Stat())

	fmt.Println("Cleaning database...")
	err = cleanDatabase(ctx, pool)
	check(err)
	os.Remove("data/user_ids.csv")
	os.Remove("data/account_ids.csv")

	fmt.Println("Seeding sample data...")
	userIds, err := seedUsers(ctx, pool, 100, time.Now().UnixMilli())

	accountIds, err := seedAccounts(ctx, pool, userIds, 250, time.Now().UnixMilli())
	check(err)

	// err = seedTransactions(ctx, pool, accountIds, 1000000, time.Now().UnixMilli())
	// if err != nil {
	// 	fmt.Println("Failed to seed transactions", "error", err)
	// 	return
	// }

	fmt.Println("Sample data inserted successfully")

	os.Mkdir("data", 0755)
	err = exportIDsToCSV(userIds, "data/user_ids.csv")
	check(err)
	err = exportIDsToCSV(accountIds, "data/account_ids.csv")
	check(err)

	fmt.Println("User and account IDs exported to CSV files successfully")
}
