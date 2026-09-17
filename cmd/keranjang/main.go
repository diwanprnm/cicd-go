// Command keranjang menjalankan HTTP API penghitung belanja.
//
// Variabel version dan commit di bawah diisi oleh pipeline lewat flag
// linker saat build:
//
//	go build -ldflags "-X main.version=1.2.3 -X main.commit=abc1234"
//
// Kalau di-build manual tanpa flag itu, nilainya tetap "dev" dan "none".
// Ini cara pipeline "menstempel" binary dengan identitas versinya.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/diwan/keranjang/internal/api"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	addr := flag.String("addr", ":8080", "alamat listen HTTP")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	logger.Info("keranjang mulai", "version", version, "commit", commit, "addr", *addr)

	srv := &http.Server{
		Addr:              *addr,
		Handler:           api.NewServer(logger).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Server dijalankan di goroutine terpisah supaya main bisa menunggu
	// sinyal berhenti.
	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	// Graceful shutdown. Di container ini penting: saat pipeline memindahkan
	// traffic ke versi baru, container lama menerima SIGTERM. Tanpa ini,
	// request yang sedang berjalan akan terputus paksa.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		logger.Error("server gagal listen", "error", err)
		os.Exit(1)
	case <-ctx.Done():
		logger.Info("sinyal berhenti diterima, menutup server dengan rapi")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("gagal shutdown dengan rapi", "error", err)
		os.Exit(1)
	}
}
