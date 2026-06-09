#!/bin/bash
# Chaos test for backend-billing -- the structural mirror of the
# kube-billing chaos test in kube-billing/bench/results-chaos/.
#
# Outputs (in $OUT_DIR):
#   requests.csv          -- t_ms,phase,code  (one line per attempt)
#   chaos-events.csv      -- t_ms,event_label

set -uo pipefail

OUT_DIR="${OUT_DIR:-/tmp/chaos-backend}"
mkdir -p "$OUT_DIR"
REQ_CSV="$OUT_DIR/requests.csv"
EV_CSV="$OUT_DIR/chaos-events.csv"
LOG="$OUT_DIR/chaos.log"

API="${API:-http://localhost:8080}"
PLAN="${PLAN:-basic}"
PHASE_SUBS=30
PHASE_C=5
DOWN_SECONDS=12

CONTAINER="${CONTAINER:-backend-billing-api-1}"

echo "t_ms,phase,code" > "$REQ_CSV"
echo "t_ms,event" > "$EV_CSV"
: > "$REQ_CSV.lock"

t_start_ms=$(($(date +%s%N) / 1000000))

now_ms() {
    echo $(( ($(date +%s%N) / 1000000) - t_start_ms ))
}

log_event() {
    echo "$(now_ms),$1" >> "$EV_CSV"
    echo "[$(now_ms) ms] $1" >&2
}

send_batch() {
    local prefix=$1
    local n=$2
    local c=$3
    log_event "batch_start_${prefix}_n${n}_c${c}"
    export REQ_CSV API PLAN
    seq 1 "$n" | xargs -n 1 -P "$c" -I {} bash -c '
        i=$1
        prefix=$2
        payload="{\"userId\":\"chaos-${prefix}-${i}\",\"planRef\":\"${PLAN}\"}"
        code=$(curl -sS -o /dev/null -m 2 -w "%{http_code}" \
            -H "Content-Type: application/json" -d "$payload" \
            "${API}/subscriptions" 2>/dev/null || echo "000")
        t=$(( $(date +%s%N) / 1000000 ))
        (
            flock 9
            echo "${t},${prefix},${code}" >> "$REQ_CSV"
        ) 9>>"${REQ_CSV}.lock"
    ' _ {} "$prefix" >> "$LOG" 2>&1
    log_event "batch_end_${prefix}"
}

kill_api() {
    log_event "api_kill_requested"
    docker kill "$CONTAINER" >> "$LOG" 2>&1
    log_event "api_killed"
}

restart_api() {
    log_event "api_restart_requested"
    docker start "$CONTAINER" >> "$LOG" 2>&1
    until curl -sS -m 2 "$API/ready" 2>/dev/null | grep -q ok; do
        sleep 0.5
    done
    log_event "api_ready"
}

# --- Run ---
log_event "experiment_start"

send_batch p1 "$PHASE_SUBS" "$PHASE_C"
sleep 5

kill_api
send_batch p2 "$PHASE_SUBS" "$PHASE_C"
sleep "$DOWN_SECONDS"

restart_api
send_batch p3 "$PHASE_SUBS" "$PHASE_C"
sleep 5

log_event "experiment_end"

# Rebase request timestamps to experiment time.
python3 - "$REQ_CSV" "$t_start_ms" <<'PY'
import csv, sys
path, t0 = sys.argv[1], int(sys.argv[2])
rows = list(csv.DictReader(open(path)))
with open(path, "w") as f:
    w = csv.writer(f)
    w.writerow(["t_ms", "phase", "code"])
    for r in rows:
        w.writerow([int(r["t_ms"]) - t0, r["phase"], r["code"]])
PY

python3 - "$REQ_CSV" <<'PY'
import csv, sys, collections
counts = collections.Counter()
phase_codes = collections.defaultdict(collections.Counter)
for r in csv.DictReader(open(sys.argv[1])):
    counts[r["code"]] += 1
    phase_codes[r["phase"]][r["code"]] += 1
print("by code:", dict(counts))
for phase, cnt in sorted(phase_codes.items()):
    print(f"  {phase}: {dict(cnt)}")
PY

echo "requests:   $REQ_CSV"
echo "events:     $EV_CSV"
