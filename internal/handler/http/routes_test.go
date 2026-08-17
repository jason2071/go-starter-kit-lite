package http

import (
	"io"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthRoute(t *testing.T) {
	app := NewApp(Dependencies{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})

	response, err := app.Test(httptest.NewRequest(stdhttp.MethodGet, "/healthz", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, stdhttp.StatusOK)
	}
}
