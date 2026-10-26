package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hotel-app/routes"
)

// TestHealthCheckEndpoint menguji respons status endpoint /api/health
func TestHealthCheckEndpoint(t *testing.T) {
	router := routes.SetupRouter()

	req, err := http.NewRequest("GET", "/api/health", nil)
	if err != nil {
		t.Fatalf("Gagal membuat HTTP request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ekspektasi HTTP 200, didapatkan %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Gagal membaca respons JSON: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Ekspektasi status 'ok', didapatkan '%v'", response["status"])
	}

	if response["app"] != "Aplikasi Akuntansi & Keuangan Bisnis Hotel" {
		t.Errorf("Nama aplikasi tidak sesuai: %v", response["app"])
	}
}
