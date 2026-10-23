// Command server runs the ektevalgomat HTTP server. Single static binary,
// no config files, no database — the only inputs are two optional env
// vars (PORT, defaults to 8080; ADDR, defaults to all interfaces).
//
// ektevalgomat — a sourced, evidence-based Norwegian election guide.
// Copyright (C) 2026  Jon-Mikkel Korsvik
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public
// License along with this program. If not, see
// <https://www.gnu.org/licenses/>. Full text: ./LICENSE.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jkorsvik/ektevalgomat/internal/server"
)

// A scratch image has no shell, no wget, no curl — nothing Docker's
// classic `HEALTHCHECK CMD wget ...` could run. The binary checks itself;
// docker-compose.yml's healthcheck runs `/server healthcheck`.
func maybeRunHealthcheck() {
	if len(os.Args) < 2 || os.Args[1] != "healthcheck" {
		return
	}
	port := envOr("PORT", "8080")
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + port + "/healthz")
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
	os.Exit(0)
}

func main() {
	maybeRunHealthcheck()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	addr := envOr("ADDR", "") + ":" + envOr("PORT", "8080")

	srv := &http.Server{
		Addr:              addr,
		Handler:           server.New(logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown", "err", err)
	}
}

func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
