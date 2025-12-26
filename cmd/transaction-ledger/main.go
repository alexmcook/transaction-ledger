package main

import (
	"context"
	"os"
	"fmt"
	"net/http"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	connStr, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		fmt.Println("DATABASE_URL not set in environment")
		return
	}

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		fmt.Printf("Unable to create connection pool: %v\n", err)
		return
	}
	defer pool.Close()

	var t1, t2 string
	rows, err := pool.Query(context.Background(), "SELECT * FROM transaction_types")
	if err != nil {
		fmt.Printf("QueryRow failed: %v\n", err)
		return
	}
	for rows.Next() {
		if err := rows.Scan(&t1, &t2); err != nil {
			fmt.Printf("Row scan failed: %v\n", err)
			return
		}
		fmt.Println(t1, t2)
	}


	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

	})
	
	http.ListenAndServe(":8080", nil)
}
