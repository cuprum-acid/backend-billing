#!/bin/bash
# Resource-utilisation sampling for backend-billing (§5.4 of the thesis).
# Phases: idle 120s -> steady (loadgen -rps 8) -> saturated (c=20) -> 3x restart.
# Outputs: resources.csv (epoch_s,container,cpu_pct,mem_MiB),
#          events.csv (epoch_s,event), startup.csv (trial,startup_s).
set -e
cd "$(dirname "$0")"
RES=resources.csv; EV=events.csv; SU=startup.csv
echo "epoch_s,container,cpu_pct,mem_MiB" > $RES
echo "epoch_s,event" > $EV
echo "trial,startup_s" > $SU
ev() { echo "$(date +%s),$1" >> $EV; }

( while true; do
    TS=$(date +%s)
    docker stats --no-stream --format '{{.Name}},{{.CPUPerc}},{{.MemUsage}}' 2>/dev/null \
      | awk -F, -v ts=$TS '{
          cpu=$2; gsub(/%/,"",cpu);
          split($3,m," / "); v=m[1];
          if (v ~ /GiB/) {gsub(/GiB/,"",v); v=v*1024} else {gsub(/MiB/,"",v)}
          print ts","$1","cpu","v }' >> $RES
    sleep 3
  done ) & SAMPLER=$!
trap "kill $SAMPLER 2>/dev/null" EXIT

ev idle_start;  sleep 120;  ev idle_end

ev steady_start
( cd .. && go run . -base http://localhost:8080 -n 1440 -c 2 -rps 8 -warmup 0 -op get-subscription -plan basic ) > steady-run.log 2>&1
ev steady_end

ev saturated_start
( cd .. && go run . -base http://localhost:8080 -n 120000 -c 20 -warmup 0 -op get-subscription -plan basic ) > saturated-run.log 2>&1
ev saturated_end

kill $SAMPLER 2>/dev/null; trap - EXIT

for i in 1 2 3; do
  docker compose -f ../../docker-compose.yml restart api > /dev/null 2>&1
  T0=$(docker inspect backend-billing-api-1 --format '{{.State.StartedAt}}')
  T0S=$(python3 -c "from datetime import datetime; print(datetime.fromisoformat('$T0'.replace('Z','+00:00')).timestamp())")
  while ! curl -sf -o /dev/null http://localhost:8080/ready; do sleep 0.2; done
  T1S=$(python3 -c "import time; print(time.time())")
  echo "$i,$(python3 -c "print(round($T1S-$T0S,2))")" >> $SU
done
echo DONE
