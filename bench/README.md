# backend-billing benchmark harness

This directory contains the load generator and microbenchmarks that produce
the numbers Chapter 5 of the thesis reports for `backend-billing`.

## What is measured here

Two kinds of measurements:

- **Microbenchmarks** (`go test -bench`). These live in
  `../handlers/benchmark_test.go` and exercise pure-CPU paths (validator,
  JSON decode + validate, error response writer, healthz handler). They
  isolate the cost of the non-database portion of each handler. Run with:

  ```
  cd .. && go test -bench=. -benchmem -benchtime=2s -count=3 -run=^$ ./handlers/
  ```

- **End-to-end latency / throughput** (`loadgen.go`). A concurrent HTTP
  driver that targets the running server (Docker Compose or a local
  binary) and records per-request latency to CSV.

## Running the load generator

Start the backend and database:

```
docker-compose up --build -d
# wait until /ready returns 200
```

Then run the harness:

```
go run ./bench \
    -base http://localhost:8080 \
    -n 5000 -c 50 \
    -warmup 60s \
    -op create-subscription \
    -plan basic \
    -out subscriptions.csv
```

The harness prints p50/p95/p99 latency and sustained throughput, and writes a
CSV with one row per request (`start_ns,duration_ns,status`). The CSV is
what Chapter 5 aggregates into the latency tables.

Available operations: `create-plan`, `create-subscription`,
`get-subscription`, `cancel-subscription`. Each operation is driven
independently, so an idiomatic full sweep is:

```
make bench
```

(from the project root, which runs the four operations sequentially and
saves their CSVs under `bench/results/`).

## Reproducing the thesis numbers

The numbers in Chapter 5 were captured on:

- Apple M3 host (4 P + 4 E cores), 8 GB unified memory, macOS 15.7.7
- Docker Desktop 27.5.0 with 4 vCPU / 6 GB allocation
- backend-billing at the commit recorded in the thesis
- Each run preceded by a 60-second warm-up phase whose results are
  discarded (`-warmup 60s`)
- Each run repeated 10 times; percentiles aggregated across the retained
  runs

The raw CSVs that underlie the reported numbers are not committed (they
are reproducible from this harness); to regenerate them on a clean host:

```
make clean
docker-compose up -d
sleep 30  # wait for /ready
for op in create-plan create-subscription get-subscription cancel-subscription; do
    for i in $(seq 1 10); do
        go run ./bench -base http://localhost:8080 -n 5000 -c 50 \
            -warmup 60s -op $op -out bench/results/${op}-run$i.csv
    done
done
```
