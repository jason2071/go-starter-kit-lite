package main

import (
	"io"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	httpHandler "github.com/jason2071/go-starter-kit-lite/internal/handler"
)

func TestRegisterRoutesAddsHealthEndpoint(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	dependencies := httpHandler.Dependencies{}
	app := newApp(logger, "*")
	registerRoutes(app, httpHandler.NewHandler(dependencies), nil)

	response, err := app.Test(httptest.NewRequest(stdhttp.MethodGet, "/healthz", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	if response.StatusCode != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, stdhttp.StatusOK)
	}
}
