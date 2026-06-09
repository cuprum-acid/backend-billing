# backend-billing chaos test — 2026-06-05

Structural mirror of the kube-billing chaos test in
`../../kube-billing/bench/results-chaos/`. The experimental design
is deliberately parallel so the resulting figure can be compared
directly.

## Procedure (three phases)

1. **API alive.** docker-compose stack is up; PostgreSQL holds a
   single seeded `basic` BillingPlan. The driver sends 30
   `POST /subscriptions` requests at concurrency 5; all return 201.
2. **API killed.** `docker kill backend-billing-api-1`. The driver
   sends 30 more `POST /subscriptions` requests; every one fails
   at the HTTP layer (curl reports `000`, the no-response sentinel).
   Crucially, *the failed requests never reach durable storage* --
   PostgreSQL is still up, but the API process between it and the
   client is gone, so the requests are simply rejected.
3. **API restarted.** `docker start backend-billing-api-1`, wait
   for `/ready`. The driver sends 30 more requests; all return 201.

The bash driver uses `flock` to serialise the append into
`requests.csv` so parallel `xargs` workers do not race on the log
file (the lock-file is `requests.csv.lock`). Each line in the
output is one attempted request: timestamp (relative to experiment
start), phase label, HTTP code (or `000` for "no response").

## Files

| File | Schema |
| --- | --- |
| `chaos.sh` | Driver script. Parameterised on `$OUT_DIR`, `$API`, `$CONTAINER`. |
| `requests.csv` | `t_ms,phase,code` (one row per attempt) |
| `chaos-events.csv` | `t_ms,event` (api_kill_requested, api_killed, api_restart_requested, api_ready, batch_start/end_p*) |

## Headline findings

- **All 30 phase-2 requests fail at the HTTP layer.** No row reaches
  PostgreSQL. There is no "backlog" to drain after restart, because
  there is nothing for a recovering process to drain from.
- **Total successes after the experiment: 60.** The phase-2
  intentions are simply lost. A client that needs durability has to
  retry on its own; the backend offers no such guarantee.
- **Contrast with kube-billing.** Under the same three-phase
  procedure but with the controller killed instead of the API, the
  phase-2 subscriptions *land in etcd* (the apiserver is a
  separate process from the controller) and are reconciled to
  Active in 1.0 s after the controller restarts. Total successes
  after that experiment: 60, with no client-side retry.

The point is structural rather than implementational: the operator
pattern decouples *accept-write* from *do-work*; the REST pattern
does not. The §5.7 consistency-models discussion frames this as a
trade-off; the two paired figures (this one + the kube one) turn
the trade-off into a measurement.
