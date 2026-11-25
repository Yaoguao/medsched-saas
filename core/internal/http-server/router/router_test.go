package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"base/internal/repository"
	"base/internal/service"
)

func TestRouterHealth(t *testing.T) {
	t.Parallel()

	r := New(service.New(repository.NewStore(nil)), "test", "")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}
}
