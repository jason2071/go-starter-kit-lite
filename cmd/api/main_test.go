package main

import (
	"io"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	httpHandler "github.com/jason2071/go-starter-kit-lite/internal/handler/http"
)

func TestRegisterRoutesAddsHealthEndpoint(t *testing.T) {
	dependencies := httpHandler.Dependencies{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	app := httpHandler.NewApp(dependencies)
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
