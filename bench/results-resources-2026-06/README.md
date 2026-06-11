# backend-billing resource-utilisation sampling — 2026-06-11

Raw resource-utilisation timeseries backing §5.4 of the thesis, replacing
the earlier interactive estimates with committed, reproducible data.
Captured on the reference Apple M3 host (8 GB unified memory,
macOS 15.7.7) against the docker-compose stack (PostgreSQL 15-alpine,
Jaeger, Prometheus) — same host as `results-2026-06/`.

## Procedure (`sample-resources.sh`)

A background sampler records `docker stats --no-stream` for every
container roughly every 3–5 s (the `docker stats` call itself takes
~2 s). Phases, delimited in `events.csv`:

1. **idle** — 120 s with no client traffic.
2. **steady** — `loadgen -rps 8 -c 2 -n 1440 -op get-subscription`
   (the same rate-capped regime as the latency tables; ~3 min).
3. **saturated** — `loadgen -c 20 -n 120000 -op get-subscription`
   with no client cap (rate limiter saturated, ~18 % 429s; ~1 min).
4. **startup ×3** — `docker compose restart api`; startup time is
   measured from the container's `State.StartedAt` to the first
   successful `GET /ready`, polled every 200 ms.

## Files

| File | Schema |
| --- | --- |
| `resources.csv` | `epoch_s,container,cpu_pct,mem_MiB` (cpu_pct is percent of one core, as reported by `docker stats`; millicores ≈ cpu_pct × 10) |
| `events.csv` | `epoch_s,event` phase boundaries |
| `startup.csv` | `trial,startup_s` |
| `steady-run.log`, `saturated-run.log` | loadgen console output for the load phases |

## Caveats

- `docker stats` samples are instantaneous readings ~3–5 s apart, not
  averages over the interval; the thesis reports per-phase medians.
- The `minikube` container shares the Docker VM and appears in the
  CSV; it is excluded from the backend analysis.
