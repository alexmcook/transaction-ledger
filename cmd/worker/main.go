package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alexmcook/transaction-ledger/internal/logger"
	"github.com/alexmcook/transaction-ledger/internal/worker"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
)

func ensureTopicExists(ctx context.Context, client *kgo.Client, topic string) error {
	adm := kadm.NewClient(client)

	topics, err := adm.ListTopics(ctx, topic)
	if err != nil {
		return fmt.Errorf("failed to list topics: %v", err)
	}

	if topics.Has(topic) {
		return nil
	}

	if _, err := adm.CreateTopics(ctx, 64, 1, nil, topic); err != nil {
		return fmt.Errorf("failed to create topic: %v", err)
	}

	return nil
}

func getPartitionOffsets(pool *pgxpool.Pool, minPart int, maxPart int) (map[string]map[int32]kgo.Offset, error) {
	assignments := make(map[int32]kgo.Offset)
	for i := minPart; i <= maxPart; i++ {
		assignments[int32(i)] = kgo.NewOffset().AtStart()
	}

	const query = `SELECT partition_id, last_offset FROM kafka_offsets WHERE partition_id BETWEEN $1 AND $2`
	rows, err := pool.Query(context.Background(), query, minPart, maxPart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var partitionID int32
		var offset int64
		if err := rows.Scan(&partitionID, &offset); err != nil {
			return nil, err
		}
		assignments[partitionID] = kgo.NewOffset().At(offset + 1)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return map[string]map[int32]kgo.Offset{"transactions": assignments}, nil
}

func setup(minPart int, maxPart int) (worker.WriterInterface, func(), error) {
	var closures []func()
	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			for i := len(closures) - 1; i >= 0; i-- {
				closures[i]()
			}
		})
	}

	log := logger.NewLogger(slog.LevelInfo)
	log.Info(fmt.Sprintf("Starting transaction ledger worker for partitions %d-%d", minPart, maxPart))

	dbUrl, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		return nil, cleanup, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to connect to database: %v", err)
	}
	closures = append(closures, pool.Close)

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = pool.Ping(pingCtx)
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to ping database: %v", err)
	}

	topicPartitions, err := getPartitionOffsets(pool, minPart, maxPart)
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to get partition offsets: %v", err)
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(os.Getenv("KAFKA_BROKERS")),
		kgo.ConsumePartitions(topicPartitions),
	)
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to create broker client: %v", err)
	}
	closures = append(closures, client.Close)

	pingCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = client.Ping(pingCtx)
	if err != nil {
		return nil, cleanup, fmt.Errorf("failed to ping broker client: %v", err)
	}

	topicCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = ensureTopicExists(topicCtx, client, "transactions")

	writer := worker.NewEfficientWriter(log, pool, client)

	return writer, cleanup, nil
}

func parsePartitionRange(partitionRange string) (int, int, error) {
	var minPartition, maxPartition int
	_, err := fmt.Sscanf(partitionRange, "%d-%d", &minPartition, &maxPartition)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid partition range format")
	}
	if minPartition < 0 || maxPartition < minPartition {
		return 0, 0, fmt.Errorf("invalid partition range values")
	}
	return minPartition, maxPartition, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	partitionRange := flag.String("partitions", "0-63", "Range of partitions to consume, e.g. '0-3'")
	flag.Parse()

	minPartition, maxPartition, err := parsePartitionRange(*partitionRange)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid partition range: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Worker starting on partitions: %s\n", *partitionRange)

	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start pprof server: %v\n", err)
			os.Exit(1)
		}
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8080", nil)
	}()

	writer, cleanup, err := setup(minPartition, maxPartition)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set up server: %v\n", err)
		cleanup()
		os.Exit(1)
	}

	go func() {
		if err := writer.Start(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start worker: %v\n", err)
			stop()
		}
	}()

	<-ctx.Done()
	fmt.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := writer.Stop(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop worker: %v\n", err)
	}

	cleanup()
	fmt.Println("Server stopped")
}
