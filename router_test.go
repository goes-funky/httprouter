package httprouter_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goes-funky/httprouter"
	"github.com/goes-funky/zapdriver"
	"github.com/google/go-cmp/cmp"
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
		name             string
		method, path     string
		expectedStatus   int
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
			if test.expectedStatus != httpResp.Code {
				t.Errorf("expected status code %d, got %d", http.StatusNotFound, httpResp.Code)
			}

			if test.expectedResponse == nil {
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

			if diff := cmp.Diff(test.expectedResponse, &errorResp); diff != "" {
				t.Error("expected response", diff)
			}
		})
	}
}

func TestRouterOptionsHandler(t *testing.T) {

}
