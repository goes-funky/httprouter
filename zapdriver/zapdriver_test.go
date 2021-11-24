package zapdriver_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/goes-funky/httprouter"
	"github.com/goes-funky/httprouter/zapdriver"
)

func TestLogRoundtrip(t *testing.T) {
	config := zapdriver.NewDevelopmentConfig()

	core, logs := observer.New(zap.DebugLevel)
	logger, err := config.Build(zap.WrapCore(func(original zapcore.Core) zapcore.Core {
		return zapcore.NewTee(original, core)
	}))

	if err != nil {
		t.Fatal(err)
	}

	router := httprouter.New(httprouter.WithLogRoundtrip(zapdriver.LogRoundtrip(logger)))

	router.Handler(http.MethodGet, "/hello", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprint(w, "Hello World!")
		return nil
	})

	server := httptest.NewServer(router)

	client := server.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "get exisiting",
			method:         http.MethodGet,
			path:           "/hello",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "redirect",
			method:         http.MethodGet,
			path:           "/hello/",
			expectedStatus: http.StatusPermanentRedirect,
		},
		{
			name:           "not found",
			method:         http.MethodGet,
			path:           "/unknown",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(test.method, server.URL+test.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if test.expectedStatus != resp.StatusCode {
				t.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
			}

			entries := logs.TakeAll()
			if len(entries) != 1 {
				t.Fatalf("expected single log entry, got %d entries", logs.Len())
			}

			entry := entries[0]
			if "roundtrip" != entry.Message {
				t.Errorf("expected message %q, got %q", "roundtrip", entry.Message)
			}

			rawPayload := entry.ContextMap()["httpRequest"]
			if rawPayload == nil {
				t.Fatal("expected payload")
			}

			payload, ok := rawPayload.(map[string]interface{})
			if !ok {
				t.Fatal("invalid payload type")
			}

			requiredEntries := []string{"latency", "remoteIp", "userAgent", "responseSize"}
			for _, key := range requiredEntries {
				if v, ok := payload[key].(string); !ok || len(v) == 0 {
					t.Errorf("expected %q to be non empty string value", key)
				}

				delete(payload, key)
			}

			expectedPayload := map[string]interface{}{
				"protocol":      "HTTP/1.1",
				"requestMethod": test.method,
				"requestUrl":    test.path,
				"status":        test.expectedStatus,
			}

			if diff := cmp.Diff(expectedPayload, payload); diff != "" {
				t.Errorf("expected payload: %s", diff)
			}
		})
	}
}
