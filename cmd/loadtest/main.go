// Command loadtest fires OTLP log batches at the ingest API concurrently to
// measure throughput. Each HTTP request carries a batch of records (as real OTLP
// SDKs do), so a modest request rate yields a high event rate. Records cycle
// through a configurable number of distinct exception types to exercise issue
// creation, not just a single group.
//
//	go run ./cmd/loadtest -url http://localhost:8080 -key <API_KEY> -public-id <ID> \
//	    -events 100000 -batch 100 -concurrency 64 -distinct 50
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	var (
		url         = flag.String("url", "http://localhost:8080", "ingest base URL")
		key         = flag.String("key", "", "API key")
		publicID    = flag.String("public-id", "", "service public id")
		totalEvents = flag.Int("events", 100000, "total log records to send")
		batch       = flag.Int("batch", 100, "records per HTTP request")
		concurrency = flag.Int("concurrency", 64, "concurrent senders")
		distinct    = flag.Int("distinct", 50, "distinct exception types (issues)")
	)
	flag.Parse()

	if *key == "" || *publicID == "" {
		log.Fatal("-key and -public-id are required")
	}

	endpoint := fmt.Sprintf("%s/api/v1/logs/%s", *url, *publicID)
	requests := (*totalEvents + *batch - 1) / *batch

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        *concurrency * 2,
			MaxIdleConnsPerHost: *concurrency * 2,
			MaxConnsPerHost:     *concurrency * 2,
		},
	}

	// Pre-build one payload per distinct fingerprint; senders round-robin them.
	payloads := make([][]byte, *distinct)
	for i := range payloads {
		payloads[i] = buildBatch(*batch, i)
	}

	var (
		sent, failed atomic.Int64
		reqIdx       atomic.Int64
	)

	log.Printf("sending %d events (%d requests × %d records) across %d senders, %d distinct issues",
		*totalEvents, requests, *batch, *concurrency, *distinct)

	start := time.Now()
	var wg sync.WaitGroup
	for w := 0; w < *concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				n := reqIdx.Add(1) - 1
				if int(n) >= requests {
					return
				}
				body := payloads[int(n)%*distinct]
				if postBatch(client, endpoint, *key, body) {
					sent.Add(int64(*batch))
				} else {
					failed.Add(int64(*batch))
				}
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	rate := float64(sent.Load()) / elapsed.Seconds()
	log.Printf("done in %s", elapsed.Round(time.Millisecond))
	log.Printf("accepted: %d   failed: %d", sent.Load(), failed.Load())
	log.Printf("ingest throughput: %.0f events/sec", rate)
}

func postBatch(client *http.Client, endpoint, key string, body []byte) bool {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", key)

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	// Drain the body before closing so net/http can reuse the keep-alive
	// connection; otherwise every request opens a new socket and exhausts
	// ephemeral ports under high request counts.
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp.StatusCode == http.StatusAccepted
}

// buildBatch constructs an OTLP request with n error records for exception type
// variant `v`, so different variants produce different fingerprints/issues.
func buildBatch(n, v int) []byte {
	var b bytes.Buffer
	b.WriteString(`{"resourceLogs":[{"resource":{"attributes":[{"key":"deployment.environment","value":{"stringValue":"production"}},{"key":"service.version","value":{"stringValue":"v1.0.0"}}]},"scopeLogs":[{"logRecords":[`)
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b,
			`{"severityNumber":17,"severityText":"ERROR","body":{"stringValue":"load test"},"attributes":[{"key":"exception.type","value":{"stringValue":"LoadError%d"}},{"key":"exception.message","value":{"stringValue":"boom"}},{"key":"user.id","value":{"stringValue":"u-%d"}},{"key":"session.id","value":{"stringValue":"s-%d"}}]}`,
			v, i%1000, i%500)
	}
	b.WriteString(`]}]}]}`)
	return b.Bytes()
}
