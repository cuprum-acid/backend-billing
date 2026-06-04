// Command loadgen drives the backend-billing REST API at a configurable
// concurrency level and records per-request latency to a CSV file.
//
// It is the harness the thesis evaluation chapter uses to measure
// end-to-end latency and throughput for backend-billing. The companion
// harness for kube-billing lives in ../../kube-billing/bench.
//
// Usage:
//
//	go run ./bench -base http://localhost:8080 -n 5000 -c 50 \
//	     -op create-subscription -plan basic -out latencies.csv
package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type op struct {
	name   string
	method string
	path   func(i int) string
	body   func(i int) []byte
}

func main() {
	base := flag.String("base", "http://localhost:8080", "base URL of the backend-billing server")
	n := flag.Int("n", 5000, "total number of requests")
	c := flag.Int("c", 50, "concurrency")
	opName := flag.String("op", "create-subscription", "operation: create-plan|create-subscription|get-subscription|cancel-subscription")
	planRef := flag.String("plan", "basic", "plan name used for create-subscription")
	out := flag.String("out", "", "optional path to write per-request CSV (start_ns,duration_ns,status)")
	warmup := flag.Duration("warmup", 0, "discard requests issued during this initial window")
	flag.Parse()

	o, err := pickOp(*opName, *planRef)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	type sample struct {
		startNs  int64
		duration time.Duration
		status   int
	}
	results := make([]sample, *n)
	var errCount, okCount int64

	type job struct{ i int }
	jobs := make(chan job, *n)
	for i := 0; i < *n; i++ {
		jobs <- job{i}
	}
	close(jobs)

	start := time.Now()
	cutoff := start.Add(*warmup)
	var wg sync.WaitGroup
	for w := 0; w < *c; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				body := o.body(j.i)
				req, err := http.NewRequest(o.method, *base+o.path(j.i), bytes.NewReader(body))
				if err != nil {
					atomic.AddInt64(&errCount, 1)
					continue
				}
				if body != nil {
					req.Header.Set("Content-Type", "application/json")
				}
				ts := time.Now()
				resp, err := client.Do(req)
				dur := time.Since(ts)
				if err != nil {
					atomic.AddInt64(&errCount, 1)
					continue
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				if ts.Before(cutoff) {
					continue
				}
				results[j.i] = sample{ts.UnixNano(), dur, resp.StatusCode}
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					atomic.AddInt64(&okCount, 1)
				} else {
					atomic.AddInt64(&errCount, 1)
				}
			}
		}()
	}
	wg.Wait()
	total := time.Since(start)

	durations := make([]time.Duration, 0, *n)
	for _, s := range results {
		if s.duration > 0 {
			durations = append(durations, s.duration)
		}
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })

	fmt.Printf("op=%s base=%s n=%d c=%d total=%s\n", *opName, *base, *n, *c, total.Round(time.Millisecond))
	fmt.Printf("ok=%d err=%d throughput=%.1f req/s\n", okCount, errCount, float64(okCount)/total.Seconds())
	if len(durations) > 0 {
		fmt.Printf("p50=%s p95=%s p99=%s max=%s\n",
			pctile(durations, 0.50), pctile(durations, 0.95),
			pctile(durations, 0.99), durations[len(durations)-1])
	}

	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			log.Fatal(err)
		}
		defer func() { _ = f.Close() }()
		w := csv.NewWriter(f)
		_ = w.Write([]string{"start_ns", "duration_ns", "status"})
		for _, s := range results {
			if s.duration == 0 {
				continue
			}
			_ = w.Write([]string{
				fmt.Sprintf("%d", s.startNs),
				fmt.Sprintf("%d", s.duration.Nanoseconds()),
				fmt.Sprintf("%d", s.status),
			})
		}
		w.Flush()
	}
}

func pctile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	i := int(float64(len(sorted)-1) * p)
	return sorted[i]
}

func pickOp(name, plan string) (op, error) {
	switch name {
	case "create-plan":
		return op{
			name: name, method: "POST",
			path: func(_ int) string { return "/plans" },
			body: func(i int) []byte {
				b, _ := json.Marshal(map[string]string{
					"name":          fmt.Sprintf("bench-plan-%d", i),
					"price":         "9.99",
					"currency":      "USD",
					"billingPeriod": "monthly",
				})
				return b
			},
		}, nil
	case "create-subscription":
		return op{
			name: name, method: "POST",
			path: func(_ int) string { return "/subscriptions" },
			body: func(i int) []byte {
				b, _ := json.Marshal(map[string]string{
					"userId":  fmt.Sprintf("bench-user-%d", i),
					"planRef": plan,
				})
				return b
			},
		}, nil
	case "get-subscription":
		return op{
			name: name, method: "GET",
			path: func(_ int) string { return "/subscriptions/1" },
			body: func(_ int) []byte { return nil },
		}, nil
	case "cancel-subscription":
		return op{
			name: name, method: "POST",
			path: func(i int) string { return fmt.Sprintf("/subscriptions/%d/cancel", i+1) },
			body: func(_ int) []byte { return nil },
		}, nil
	default:
		return op{}, fmt.Errorf("unknown op: %s", name)
	}
}
