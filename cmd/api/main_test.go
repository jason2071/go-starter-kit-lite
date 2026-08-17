package main

import (
	"io"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/jason2071/go-starter-kit-lite/internal/platform"
)

func TestRegisterRoutesAddsHealthEndpoint(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := newApp(logger, "*")
	registerRoutes(app, routeDependencies{
		Ready: func() error { return nil },
		Config: platform.Config{
			JWTSecret:   "test-secret",
			JWTIssuer:   "test",
			JWTAudience: "test",
		},
	})

	response, err := app.Test(httptest.NewRequest(stdhttp.MethodGet, "/healthz", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, stdhttp.StatusOK)
	}
}
