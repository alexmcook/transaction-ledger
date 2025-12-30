#!/bin/bash
echo "Starting benchmark..."
docker compose down && docker compose up -d --build
echo "Waiting for services to be ready..."
sleep 5
echo "Seeding data..."
go run cmd/seeder/main.go
sleep 2
echo "Running stress test..."
k6 run stress_test.js 2>&1 | tee logs/benchmark_$(date +%Y%m%d_%H%M%S).log
echo "Benchmark completed"
