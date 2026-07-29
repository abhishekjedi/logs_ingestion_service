package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type demoConfig struct {
	Addr      string
	IngestURL string
	APIKey    string
}

func main() {
	cfg := demoConfig{
		Addr:      env("REPLAY_DEMO_ADDR", ":8081"),
		IngestURL: os.Getenv("ERRLOG_INGEST_URL"),
		APIKey:    os.Getenv("ERRLOG_API_KEY"),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/checkout", checkoutHandler(cfg))

	log.Printf("replay demo backend listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, mux); err != nil {
		log.Fatal(err)
	}
}

func checkoutHandler(cfg demoConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if cfg.IngestURL == "" || cfg.APIKey == "" {
			http.Error(w, "configure ERRLOG_INGEST_URL and ERRLOG_API_KEY", http.StatusServiceUnavailable)
			return
		}

		err := publishError(r, cfg)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err != nil {
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "payment failed; telemetry publish failed"})
			log.Printf("publish demo error: %v", err)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "payment provider timed out"})
	}
}

func publishError(source *http.Request, cfg demoConfig) error {
	now := time.Now().UTC()
	attributes := []map[string]any{
		otelAttribute("exception.type", "StripeTimeoutError"),
		otelAttribute("exception.message", "payment provider timed out after 500ms"),
		otelAttribute("exception.stacktrace", "StripeTimeoutError: payment provider timed out\n    at submitPayment (src/routes/checkout.tsx:88)\n    at checkout (src/api/payment.ts:41)"),
		otelAttribute("user.id", "demo-user-918"),
		otelAttribute("session.id", source.Header.Get("X-OpenReplay-Session-ID")),
		otelAttribute("openreplay.project.key", source.Header.Get("X-OpenReplay-Project-Key")),
		otelAttribute("openreplay.session.url", source.Header.Get("X-OpenReplay-Session-URL")),
		otelAttribute("http.request.method", "POST"),
		otelAttribute("url.path", "/checkout"),
		otelAttribute("http.response.status_code", "500"),
	}
	payload := map[string]any{
		"resourceLogs": []map[string]any{{
			"resource": map[string]any{"attributes": []map[string]any{
				otelAttribute("service.name", "checkout-demo"),
				otelAttribute("service.version", "demo-v1"),
				otelAttribute("deployment.environment", "local"),
			}},
			"scopeLogs": []map[string]any{{
				"scope": map[string]string{"name": "replay-demo"},
				"logRecords": []map[string]any{{
					"timeUnixNano":         fmt.Sprintf("%d", now.UnixNano()),
					"observedTimeUnixNano": fmt.Sprintf("%d", now.UnixNano()),
					"severityNumber":       17,
					"severityText":         "ERROR",
					"body":                 map[string]any{"stringValue": "Checkout failed: payment provider timed out"},
					"attributes":           attributes,
					"traceId":              "8af2a6b8e7c8402aa3dc78de304e5321",
					"spanId":               "2d45f7a18c452d91",
				}},
			}},
		}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPost, cfg.IngestURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", cfg.APIKey)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return fmt.Errorf("ingest returned %d: %s", response.StatusCode, message)
	}
	return nil
}

func otelAttribute(key, value string) map[string]any {
	return map[string]any{"key": key, "value": map[string]string{"stringValue": value}}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
