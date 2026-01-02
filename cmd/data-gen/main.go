package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"

	pb "github.com/alexmcook/transaction-ledger/api/proto/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func makeBatch(records [][]string, i int) {
	batch := &pb.TransactionBatch{
		Transactions: make([]*pb.CreateTransactionRequest, 0, 1000),
	}

	for i := range 1000 {
		id, _ := uuid.Parse(records[rand.Intn(len(records))][0])
		idBytes, _ := id.MarshalBinary()

		batch.Transactions = append(batch.Transactions, &pb.CreateTransactionRequest{
			AccountId:       idBytes,
			TransactionType: 1,
			Amount:          int64(i * 10),
		})
	}

	out, err := proto.Marshal(batch)
	if err != nil {
		log.Fatalf("Failed to marshal batch: %v", err)
	}

	err = os.WriteFile(fmt.Sprintf("data/tx/tx_batch_%d.bin", i), out, 0644)
	if err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	log.Printf("Successfully wrote tx_batch_%d.bin", i)
}

func main() {
	f, err := os.Open("data/account_ids.csv")
	if err != nil {
		log.Fatalf("Failed to open account_ids.csv: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read account_ids.csv: %v", err)
	}

	for i := range 100 {
		makeBatch(records, i)
	}
}
