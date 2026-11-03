package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestServerHealthAndHeadlessRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := buildRouter()

	t.Run("GET /live returns 200 JSON ok", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/live", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		var resp map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode json: %v", err)
		}
		if resp["status"] != "ok" {
			t.Fatalf("expected status ok, got %v", resp["status"])
		}
	})

	t.Run("GET /unknown returns 404 JSON not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/unknown", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
		var resp map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode json (got non-JSON or HTML): %v, body: %s", err, w.Body.String())
		}
		if resp["error"] != "Not found" {
			t.Fatalf("expected error 'Not found', got %v", resp["error"])
		}
	})

	t.Run("GET /map returns 404 JSON not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/map", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404 for removed static /map, got %d", w.Code)
		}
	})

	t.Run("GET /favicon.ico returns 404 JSON not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/favicon.ico", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404 for removed favicon, got %d", w.Code)
		}
	})

	t.Run("GET /ready returns JSON status", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ready", nil)
		router.ServeHTTP(w, req)

		// When databases are not running, it must return 503 unavailable
		var resp map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode json: %v, body: %s", err, w.Body.String())
		}
		if resp["status"] != "ready" && resp["status"] != "unavailable" {
			t.Fatalf("unexpected status: %v", resp["status"])
		}
	})
}
