package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/alexmcook/transaction-ledger/proto"
	"github.com/google/uuid"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func setup(n int, batchSize int) [][]byte {
	accountIDs, err := loadAccountIDs("data/account_ids.csv")
	if err != nil {
		log.Fatal("Failed to load account IDs:", err)
	}

	data := make([][]byte, n)

	for i := range n {
		batch := &pb.TransactionBatch{
			Transactions: make([]*pb.Transaction, batchSize),
		}

		seed := rand.Intn(len(accountIDs))
		for j := range batch.Transactions {
			id, _ := uuid.NewV7()
			batch.Transactions[j] = &pb.Transaction{
				Id:        id[:],
				AccountId: accountIDs[seed%len(accountIDs)],
				Amount:    int64(1000),
			}
			seed++
		}

		proto, err := batch.MarshalVT()
		if err != nil {
			log.Fatal("Marshal error:", err)
		}

		data[i] = proto
	}

	return data
}

func loadAccountIDs(filePath string) ([][]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(data, []byte{'\n'})
	accountIDs := make([][]byte, 0, len(lines))

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		id, err := uuid.ParseBytes(line)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID in account IDs file: %v", err)
		}
		accountIDs = append(accountIDs, id[:])
	}

	return accountIDs, nil
}

func main() {
	const (
		numFiles  = 100
		batchSize = 1000
		targetRPS = 500 * 1000 / batchSize
		targetURL = "http://localhost/transactions/proto"
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	batches := setup(numFiles, batchSize)
	counter := 0
	targeter := func(t *vegeta.Target) error {
		if t == nil {
			return vegeta.ErrNilTarget
		}
		t.Method = "POST"
		t.URL = targetURL
		t.Body = batches[counter%len(batches)]
		t.Header = map[string][]string{
			"Content-Type": {"application/x-protobuf"},
		}
		counter++
		return nil
	}

	rate := vegeta.Rate{Freq: targetRPS, Per: time.Second}
	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics
	duration := 30 * time.Second

	go func() {
		<-sigChan
		log.Println("Attack interrupted, stopping...")
		attacker.Stop()
	}()

	fmt.Printf("Starting attack: %d RPS to %s\n", targetRPS, targetURL)

	for res := range attacker.Attack(targeter, rate, duration, "Transaction Load Test") {
		metrics.Add(res)
	}

	metrics.Close()

	fmt.Printf("Attack complete:\n")
	fmt.Printf("  Requests: %d\n", metrics.Requests)
	fmt.Printf("  TPS: %f\n", metrics.Throughput*batchSize)
	fmt.Printf("  Success: %.2f%%\n", metrics.Success*100)
	fmt.Printf("  Latencies:\n")
	fmt.Printf("    Mean: %s\n", metrics.Latencies.Mean)
	fmt.Printf("    P95: %s\n", metrics.Latencies.P95)
	fmt.Printf("    P99: %s\n", metrics.Latencies.P99)
	fmt.Printf("    Max: %s\n", metrics.Latencies.Max)
}
