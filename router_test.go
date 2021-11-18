package httprouter_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"

	"github.com/goes-funky/httprouter"
)

func TestRouterDefaultHandlers(t *testing.T) {
	logger, err := zap.NewDevelopmentConfig().Build()
	if err != nil {
		t.Fatal(err)
	}

	router := httprouter.New(logger, httprouter.WithVerbose(true))

	router.Handler(http.MethodGet, "/", func(w http.ResponseWriter, req *http.Request) error {
		fmt.Fprint(w, "Hello World!")
		return nil
	})

	router.Handler(http.MethodGet, "/error", func(w http.ResponseWriter, req *http.Request) error {
		return httprouter.NewError(
			http.StatusForbidden,
			httprouter.Message("forbidden"),
			httprouter.Cause(errors.New("forbidden cause")),
		)
	})

	router.Handler(http.MethodGet, "/panic", func(w http.ResponseWriter, req *http.Request) error {
		panic("panic handler")
	})

	doRequest := func(method string, path string) *http.Response {
		req := httptest.NewRequest(method, "http://example.com"+path, nil)
		rw := httptest.NewRecorder()
		router.ServeHTTP(rw, req)

		return rw.Result()
	}

	tests := []struct {
		name             string
		method, path     string
		expectedStatus   int
		expectedHeaders  http.Header
		expectedResponse *httprouter.ErrorResponse
	}{
		{
			name:           "not found handler",
			method:         http.MethodGet,
			path:           "/unknown-path",
			expectedStatus: http.StatusNotFound,
			expectedResponse: &httprouter.ErrorResponse{
				Message: http.StatusText(http.StatusNotFound),
			},
		},
		{
			name:           "invalid method handler",
			method:         http.MethodHead,
			path:           "/",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "options handler",
			method: http.MethodOptions,
			path:   "/",
			expectedHeaders: http.Header{
				"Allow": []string{"GET, OPTIONS"},
			},
		},
		{
			name:           "error handler",
			method:         http.MethodGet,
			path:           "/error",
			expectedStatus: http.StatusForbidden,
			expectedResponse: &httprouter.ErrorResponse{
				Message: "forbidden",
				Debug:   "forbidden cause",
			},
		},
		{
			name:           "panic handler",
			method:         http.MethodGet,
			path:           "/panic",
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: &httprouter.ErrorResponse{
				Message: http.StatusText(http.StatusInternalServerError),
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			httpResp := doRequest(test.method, test.path)
			expectErrorResponse(t, httpResp, test.expectedStatus, test.expectedHeaders, test.expectedResponse)
		})
	}
}

func expectErrorResponse(t testing.TB, httpResp *http.Response, status int, headers http.Header, resp *httprouter.ErrorResponse) {
	if status != 0 && status != httpResp.StatusCode {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, httpResp.StatusCode)
	}

	for k, vs := range headers {
		if diff := cmp.Diff(vs, httpResp.Header[k]); diff != "" {
			t.Errorf("expected header %q %s", k, diff)
		}
	}

	if resp == nil {
		body, err := io.ReadAll(httpResp.Body)
		if err != nil {
			t.Fatal(err)
		}

		if len(body) != 0 {
			t.Error("expected no body")
		}

		return
	}

	var errorResp httprouter.ErrorResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&errorResp); err != nil {
		t.Error(err)
		return
	}

	if diff := cmp.Diff(resp, &errorResp); diff != "" {
		t.Error("expected response", diff)
	}
}
