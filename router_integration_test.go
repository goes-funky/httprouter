//go:build integration

package httprouter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/goes-funky/httprouter"
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
			name:           "options handler",
			method:         http.MethodOptions,
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedHeaders: http.Header{
				"Allow": []string{"GET, OPTIONS"},
			},
		},
	}

	for _, test := range failTests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			httpResp := doRequest(t, test.method, test.path)
			expectErrorResponse(t, httpResp, test.expectedStatus, test.expectedHeaders, test.expectedResponse)
		})
	}
}
