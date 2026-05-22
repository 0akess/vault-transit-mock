// Command vault-transit-mock starts a stateless mock of the HashiCorp
// Vault HTTP API on the configured port. It exists to remove the
// init/unseal/keyring friction of running real Vault in local
// docker-compose stacks.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"vault-transit-mock/handlers"
)

const (
	readHeaderTimeout  = 10 * time.Second
	healthcheckTimeout = 3 * time.Second
)

func main() {
	port := envDefault("PORT", "8200")

	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthcheck(port))
	}

	level := parseLevel(envDefault("LOG_LEVEL", "info"))

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	mux := http.NewServeMux()
	kv := handlers.NewKV()

	mux.HandleFunc("/v1/sys/health", handlers.Health)
	mux.HandleFunc("/v1/sys/seal-status", handlers.SealStatus)

	mux.HandleFunc("/v1/transit/keys/", handlers.TransitKeys)
	mux.HandleFunc("/v1/transit/encrypt/", handlers.TransitEncrypt)
	mux.HandleFunc("/v1/transit/decrypt/", handlers.TransitDecrypt)

	mux.HandleFunc("/v1/secret/data/", kv.Data)
	mux.HandleFunc("/v1/secret/metadata/", kv.Metadata)

	mux.HandleFunc("/v1/auth/approle/login", handlers.AppRoleLogin)
	mux.HandleFunc("/v1/auth/token/login", handlers.TokenLogin)
	mux.HandleFunc("/v1/auth/token/lookup-self", handlers.TokenLookupSelf)
	mux.HandleFunc("/v1/auth/token/renew-self", handlers.TokenRenewSelf)

	addr := ":" + port
	logger.Info("vault-transit-mock starting",
		slog.String("addr", addr),
		slog.String("version", handlers.Version),
	)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped", slog.Any("error", err))
		os.Exit(1)
	}
}

// runHealthcheck probes the running mock over loopback and returns a
// process exit code. Used as the binary's HEALTHCHECK in scratch images
// where wget/curl are unavailable.
func runHealthcheck(port string) int {
	ctx, cancel := context.WithTimeout(context.Background(), healthcheckTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:"+port+"/v1/sys/health", http.NoBody)
	if err != nil {
		return 1
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 1
	}
	//nolint:errcheck // healthcheck probe; close failure is irrelevant to the exit code
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 1
	}

	return 0
}

func envDefault(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}

	return def
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
