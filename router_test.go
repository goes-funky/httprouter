package httprouter_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goes-funky/httprouter"
	"github.com/goes-funky/zapdriver"
)

func TestRouterDefaultHandlers(t *testing.T) {
	logger, err := zapdriver.NewDevelopmentConfig().Build()
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

	doRequest := func(method string, path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "http://example.com"+path, nil)
		rw := httptest.NewRecorder()
		router.ServeHTTP(rw, req)

		return rw
	}

	tests := []struct {
		name            string
		method          string
		path            string
		expectedStatus  int
		expectedMessage string
		expectedDebug   string
	}{
		{
			name:            "not found handler",
			method:          http.MethodGet,
			path:            "/unknown-path",
			expectedStatus:  http.StatusNotFound,
			expectedMessage: http.StatusText(http.StatusNotFound),
		},
		{
			name:            "invalid method handler",
			method:          http.MethodHead,
			path:            "/",
			expectedStatus:  http.StatusMethodNotAllowed,
			expectedMessage: http.StatusText(http.StatusMethodNotAllowed),
		},
		{
			name:            "error handler",
			method:          http.MethodGet,
			path:            "/error",
			expectedStatus:  http.StatusForbidden,
			expectedMessage: "forbidden",
			expectedDebug:   "forbidden cause",
		},
		{
			name:            "panic handler",
			method:          http.MethodGet,
			path:            "/panic",
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: http.StatusText(http.StatusInternalServerError),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			res := doRequest(test.method, test.path)
			if test.expectedStatus != res.Code {
				t.Errorf("expected status code %d, got %d", http.StatusNotFound, res.Code)
			}

			var errorResp httprouter.ErrorResponse
			if err := json.NewDecoder(res.Body).Decode(&errorResp); err != nil {
				t.Error(err)
				return
			}

			if test.expectedMessage != errorResp.Message {
				t.Errorf("expected message %q, got %q", test.expectedMessage, errorResp.Message)
			}

			if test.expectedDebug != errorResp.Debug {
				t.Errorf("expected debug %q, got %q", test.expectedDebug, errorResp.Debug)
			}
		})
	}
}
