package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diwan/keranjang/internal/api"
)

func serverUji(t *testing.T) http.Handler {
	t.Helper()

	// Log dibuang supaya output test tidak berisik.
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return api.NewServer(logger).Handler()
}

func tembak(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("gagal marshal body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, reader)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	return rec
}

func TestHealthzMerespons200(t *testing.T) {
	rec := tembak(t, serverUji(t), http.MethodGet, "/healthz", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin %d", rec.Code, http.StatusOK)
	}

	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body bukan JSON valid: %v", err)
	}
	if got["status"] != "ok" {
		t.Errorf(`body["status"] = %q, ingin "ok"`, got["status"])
	}
}

func TestHitungMenghitungTotalDanDiskon(t *testing.T) {
	// 3 x 20000 = 60000 -> diskon 20% -> bayar 48000
	body := map[string]any{
		"items": []map[string]any{
			{"nama": "sepatu", "harga": 20000, "jumlah": 3},
		},
	}

	rec := tembak(t, serverUji(t), http.MethodPost, "/api/hitung", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin %d. body: %s",
			rec.Code, http.StatusOK, rec.Body)
	}

	var got struct {
		Total        float64 `json:"total"`
		PersenDiskon float64 `json:"persen_diskon"`
		Bayar        float64 `json:"bayar"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body bukan JSON valid: %v", err)
	}

	if got.Total != 60000 {
		t.Errorf("total = %.0f, ingin 60000", got.Total)
	}
	if got.PersenDiskon != 20 {
		t.Errorf("persen_diskon = %.0f, ingin 20", got.PersenDiskon)
	}
	if got.Bayar != 48000 {
		t.Errorf("bayar = %.0f, ingin 48000", got.Bayar)
	}
}

// TestHitungDiskonTepatDiBatasSeratus adalah test yang akan GAGAL
// karena bug yang sengaja ditanam di keranjang.go.
//
// Perhatikan: test ini menguji lewat HTTP, sedangkan test di package
// keranjang menguji fungsinya langsung. Dua-duanya menangkap bug yang SAMA.
// Itu memang disengaja — supaya kamu melihat bahwa satu akar masalah bisa
// muncul sebagai kegagalan di beberapa lapisan sekaligus.
func TestHitungDiskonTepatDiBatasSeratus(t *testing.T) {
	// Belanja tepat 100 -> aturannya bilang dapat diskon 10%.
	body := map[string]any{
		"items": []map[string]any{
			{"nama": "buku", "harga": 100, "jumlah": 1},
		},
	}

	rec := tembak(t, serverUji(t), http.MethodPost, "/api/hitung", body)

	var got struct {
		PersenDiskon float64 `json:"persen_diskon"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body bukan JSON valid: %v", err)
	}

	if got.PersenDiskon != 10 {
		t.Errorf("persen_diskon = %.0f, ingin 10 (total tepat 100 harus dapat diskon)",
			got.PersenDiskon)
	}
}

func TestHitungMenolakJumlahNol(t *testing.T) {
	body := map[string]any{
		"items": []map[string]any{
			{"nama": "kopi", "harga": 10000, "jumlah": 0},
		},
	}

	rec := tembak(t, serverUji(t), http.MethodPost, "/api/hitung", body)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, ingin %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHitungMenolakJSONRusak(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/hitung",
		bytes.NewReader([]byte("{bukan json")))
	rec := httptest.NewRecorder()

	serverUji(t).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, ingin %d", rec.Code, http.StatusBadRequest)
	}
}
