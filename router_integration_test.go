//go:build integration

package httprouter_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goes-funky/httprouter"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
)

type Response struct {
	Message string `json:"message"`
}

func TestIntegration(t *testing.T) {
	logger, err := zap.NewDevelopmentConfig().Build()
	if err != nil {
		t.Fatal(err)
	}

	router := httprouter.New(logger)
	router.Handler(http.MethodGet, "/", func(w http.ResponseWriter, req *http.Request) error {
		return httprouter.JSONResponse(w, http.StatusOK, Response{
			Message: "Hello World!",
		})
	})

	server := httptest.NewServer(router)

	doRequest := func(t testing.TB, method, path string) *http.Response {
		req, err := http.NewRequest(method, server.URL+path, nil)
		if err != nil {
			t.Fatal("http request", err)
		}

		resp, err := server.Client().Do(req)
		if err != nil {
			t.Fatal("http request", err)
		}

		return resp
	}

	failTests := []struct {
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
	}

	for _, test := range failTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			httpResp := doRequest(t, test.method, test.path)

			if test.expectedStatus != httpResp.StatusCode {
				t.Fatalf("expected status code %d, got %d", test.expectedStatus, httpResp.StatusCode)
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
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.expectedResponse, &errorResp); diff != "" {
				t.Error("expected response", diff)
			}
		})
	}

}
