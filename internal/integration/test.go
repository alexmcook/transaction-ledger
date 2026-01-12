package integration

//
// import (
// 	"bytes"
// 	"context"
// 	"fmt"
// 	"log"
// 	"log/slog"
// 	"net/http"
// 	"os"
// 	"testing"
// 	"time"
//
// 	"github.com/google/uuid"
// 	"github.com/jackc/pgx/v5/pgxpool"
// 	"github.com/testcontainers/testcontainers-go"
// 	"github.com/testcontainers/testcontainers-go/modules/postgres"
// 	"github.com/testcontainers/testcontainers-go/modules/redpanda"
// 	"github.com/testcontainers/testcontainers-go/wait"
// 	"github.com/twmb/franz-go/pkg/kadm"
// 	"github.com/twmb/franz-go/pkg/kgo"
//
// 	"github.com/alexmcook/transaction-ledger/internal/api"
// 	"github.com/alexmcook/transaction-ledger/internal/logger"
// 	"github.com/alexmcook/transaction-ledger/internal/storage"
// 	"github.com/alexmcook/transaction-ledger/internal/worker"
// )
//
// var (
// 	testDB     *pgxpool.Pool
// 	testBroker *kgo.Client
// )
//
// func TestMain(m *testing.M) {
// 	ctx := context.Background()
//
// 	pg, err := postgres.Run(
// 		ctx,
// 		"postgres:18-alpine",
// 		postgres.WithDatabase("testdb"),
// 		postgres.WithUsername("testuser"),
// 		postgres.WithPassword("testpass"),
// 		testcontainers.WithWaitStrategy(
// 			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(10*time.Second),
// 		),
// 	)
// 	if err != nil {
// 		log.Fatalf("Failed to start Postgres container: %v", err)
// 	}
//
// 	rp, err := redpanda.Run(
// 		ctx,
// 		"docker.redpanda.com/redpandadata/redpanda:v25.3.4",
// 		redpanda.WithAutoCreateTopics(),
// 		testcontainers.WithWaitStrategy(
// 			wait.ForLog("Started Redpanda!").WithStartupTimeout(10*time.Second),
// 		),
// 	)
// 	if err != nil {
// 		log.Fatalf("Failed to start Redpanda container: %v", err)
// 	}
//
// 	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
// 	if err != nil {
// 		log.Fatalf("Failed to get Postgres connection string: %v", err)
// 	}
//
// 	brokerAddr, err := rp.HTTPProxyAddress(ctx)
// 	if err != nil {
// 		log.Fatalf("Failed to get Redpanda broker address: %v", err)
// 	}
//
// 	testDB, err = pgxpool.New(ctx, dsn)
// 	if err != nil {
// 		log.Fatalf("Failed to connect to Postgres database: %v", err)
// 	}
//
// 	testBroker, err = kgo.NewClient(
// 		kgo.SeedBrokers(brokerAddr),
// 	)
// 	if err != nil {
// 		log.Fatalf("Failed to create Redpanda client: %v", err)
// 	}
//
// 	// Run migrations
// 	migrations, err := os.ReadFile("migrations/001_initial_schema.up.sql")
// 	if err != nil {
// 		log.Fatalf("Failed to read migrations: %v", err)
// 	}
// 	if _, err := testDB.Exec(ctx, string(migrations)); err != nil {
// 		log.Fatalf("Failed to execute migrations: %v", err)
// 	}
//
// 	// Ensure topic exists with 64 partitions
// 	adm := kadm.NewClient(testBroker)
// 	if _, err := adm.CreateTopics(ctx, 64, 1, nil, "transactions"); err != nil {
// 		log.Fatalf("Failed to create transactions topic: %v", err)
// 	}
//
// 	code := m.Run()
//
// 	testDB.Close()
// 	testBroker.Close()
// 	if err = pg.Terminate(ctx); err != nil {
// 		log.Fatalf("Failed to terminate Postgres container: %v", err)
// 	}
// 	if err = rp.Terminate(ctx); err != nil {
// 		log.Fatalf("Failed to terminate Redpanda container: %v", err)
// 	}
//
// 	os.Exit(code)
// }
//
// func TestEndToEnd(t *testing.T) {
// 	ctx := context.Background()
//
// 	// Start API server wired to test DB and broker
// 	logg := logger.NewLogger(slog.LevelDebug)
// 	shards := storage.NewShardedStore(logg, []*pgxpool.Pool{testDB})
// 	srv := api.NewServer(logg, shards, testBroker)
// 	go func() {
// 		if err := srv.Run(); err != nil {
// 			log.Fatalf("API server failed: %v", err)
// 		}
// 	}()
//
// 	// Start worker coordinator consuming all partitions
// 	coord := worker.NewCoordinator(ctx, 0, 63, logg, storage.NewPostgresStore(logg, testDB), testBroker, testDB)
// 	go func() {
// 		if err := coord.Run(ctx); err != nil {
// 			// Log and allow test to continue; coordinator may exit on client close
// 			log.Printf("Coordinator exited: %v", err)
// 		}
// 	}()
//
// 	// Wait for services to be ready
// 	time.Sleep(2 * time.Second)
//
// 	// Post a small transaction batch
// 	id := uuid.New().String()
// 	acc := uuid.New().String()
// 	payload := fmt.Sprintf(`[{"id":"%s","account_id":"%s","amount":100}]`, id, acc)
// 	resp, err := http.Post("http://localhost:8080/transactions/effjson", "application/json", bytes.NewBufferString(payload))
// 	if err != nil {
// 		t.Fatalf("Failed to POST to API: %v", err)
// 	}
// 	defer resp.Body.Close()
// 	if resp.StatusCode != 201 {
// 		t.Fatalf("Unexpected status code: %d", resp.StatusCode)
// 	}
//
// 	// Poll DB partitions for the inserted transaction
// 	accUUID, _ := uuid.Parse(acc)
// 	found := false
// 	start := time.Now()
// 	for time.Since(start) < 30*time.Second {
// 		for i := 0; i < 64; i++ {
// 			var cnt int
// 			err := testDB.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM transactions_%d WHERE account_id = $1", i), accUUID).Scan(&cnt)
// 			if err == nil && cnt > 0 {
// 				found = true
// 				break
// 			}
// 		}
// 		if found {
// 			break
// 		}
// 		time.Sleep(500 * time.Millisecond)
// 	}
// 	if !found {
// 		t.Fatalf("Timed out waiting for transaction to be written to DB")
// 	}
// }
