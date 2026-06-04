# backend-billing bench results — 2026-06-04

Raw per-request CSVs from the loadgen harness, captured on the
reference Apple M3 host (8 GB unified memory, macOS 15.7.7) against
the docker-compose stack (PostgreSQL 15 + Jaeger 1.50 + Prometheus
2.45) at backend-billing commit `33d5304`.

CSV schema: `start_ns,duration_ns,status` (one row per request, post-warmup).

## Files

| File | Regime | Invocation |
| --- | --- | --- |
| `create-plan.csv` | rps=8 steady-state | `go run ./bench -base http://localhost:8080 -n 400 -c 2 -rps 8 -warmup 10s -op create-plan -plan basic` |
| `create-subscription.csv` | rps=8 steady-state | `go run ./bench -base http://localhost:8080 -n 400 -c 2 -rps 8 -warmup 10s -op create-subscription -plan basic` |
| `get-subscription.csv` | rps=8 steady-state | `go run ./bench -base http://localhost:8080 -n 400 -c 2 -rps 8 -warmup 10s -op get-subscription -plan basic` |
| `cancel-subscription.csv` | rps=8 steady-state | `go run ./bench -base http://localhost:8080 -n 400 -c 2 -rps 8 -warmup 10s -op cancel-subscription -plan basic` |
| `create-sub-burst.csv` | burst, c=1 n=20 (fits rate-limiter budget) | `go run ./bench -n 20 -c 1 -warmup 0 -op create-subscription -plan basic` |
| `get-burst.csv` | burst, c=1 n=20 | `go run ./bench -n 20 -c 1 -warmup 0 -op get-subscription -plan basic` |
| `get-saturated.csv` | saturated, c=20 n=1000 no rate cap | `go run ./bench -n 1000 -c 20 -warmup 0 -op get-subscription -plan basic` |

## Why three regimes

The backend's rate limiter is 10 req/s + burst 20 per IP. The same
harness produces very different percentiles depending on whether the
client stays inside the budget. The thesis evaluation chapter discusses
all three explicitly so reviewers know which regime each number refers
to.

- **rps=8 steady-state** (`*.csv` per op): client-side `-rps 8` flag
  caps the bench just under the server budget, every request gets a
  token, latencies are the handler's hot path plus PostgreSQL.
- **burst** (`*-burst.csv`): 20 requests at c=1, no rate cap — fits
  inside the burst budget, no rate-limiter wait time. Reproduces the
  sub-ms p50 figure quoted in the abstract.
- **saturated** (`*-saturated.csv`): c=20 with no rate cap, ~18% of
  requests get 429s. Characterises peak throughput and the 429 tail.
