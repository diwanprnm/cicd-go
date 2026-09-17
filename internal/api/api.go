// Package api membungkus keranjang menjadi HTTP API.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/diwan/keranjang/internal/keranjang"
)

// Server menyatukan HTTP router dengan logika keranjang.
type Server struct {
	log *slog.Logger
}

// NewServer membuat Server baru.
func NewServer(log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{log: log}
}

// Handler mengembalikan http.Handler siap pakai.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("POST /api/hitung", s.handleHitung)

	return mux
}

// handleHealthz dipakai pipeline untuk memeriksa aplikasi sudah hidup.
// Endpoint sekecil ini adalah bagian penting dari pipeline: tanpa dia,
// pipeline tidak punya cara tahu aplikasi benar-benar sudah siap.
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	tulisJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type permintaanHitung struct {
	Items []keranjang.Item `json:"items"`
}

type balasanHitung struct {
	Total        float64 `json:"total"`
	PersenDiskon float64 `json:"persen_diskon"`
	Bayar        float64 `json:"bayar"`
}

func (s *Server) handleHitung(w http.ResponseWriter, r *http.Request) {
	var req permintaanHitung
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		tulisJSON(w, http.StatusBadRequest, map[string]string{
			"error": "body bukan JSON yang valid",
		})
		return
	}

	k := &keranjang.Keranjang{}
	for _, item := range req.Items {
		if err := k.Tambah(item); err != nil {
			tulisJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}
	}

	total := k.Total()
	tulisJSON(w, http.StatusOK, balasanHitung{
		Total:        total,
		PersenDiskon: keranjang.PersenDiskon(total),
		Bayar:        k.Bayar(),
	})
}

func tulisJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("gagal menulis response", "error", err)
	}
}
