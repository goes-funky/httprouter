package httprouter_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/goes-funky/httprouter"
)

type response struct {
	Message string `json:"message"`
}

func ExampleRouter() {
	router := httprouter.New()
	router.Handler(http.MethodGet, "/greet/:name", func(w http.ResponseWriter, req *http.Request) error {
		params := httprouter.GetParams(req.Context())

		response := response{Message: fmt.Sprintf("Hello %s!", params["name"])}
		return httprouter.JSONResponse(w, http.StatusOK, response)
	})

	server := httptest.NewServer(router)

	client := server.Client()

	httpResp, err := client.Get(server.URL + "/greet/fry")
	if err != nil {
		log.Fatal("failed to GET /greet/fry", err)
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		log.Fatal("failed to read http response body", err)
	}

	var resp response
	if err := json.Unmarshal(body, &resp); err != nil {
		log.Fatal("failed to unmarshal http response body", err)
	}

	if resp.Message != "Hello fry!" {
		log.Fatal("unexpected response")
	}
}

func TestRouterDefaultHandlers(t *testing.T) {
	router := httprouter.New(httprouter.WithVerbose(true))

	router.Handler(http.MethodGet, "/hello", func(w http.ResponseWriter, req *http.Request) error {
		return httprouter.JSONResponse(w, http.StatusOK, response{
			Message: "Hello World!",
		})
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

	server := httptest.NewServer(router)

	client := server.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	tests := []struct {
		name               string
		method, path       string
		ignoreResponseBody bool
		expectedStatus     int
		expectedHeaders    map[string]string
		expectedResponse   *httprouter.ErrorResponse
	}{
		{
			name:               "redirect handler",
			method:             http.MethodGet,
			path:               "/hello/",
			ignoreResponseBody: true,
			expectedStatus:     http.StatusPermanentRedirect,
			expectedHeaders: map[string]string{
				"Location": "/hello",
			},
		},
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
			path:           "/hello",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "options handler",
			method: http.MethodOptions,
			path:   "/hello",
			expectedHeaders: map[string]string{
				"Allow": "GET, OPTIONS",
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
			req, err := http.NewRequest(test.method, server.URL+test.path, nil)
			if err != nil {
				t.Fatal("http request", err)
			}

			httpResp, err := client.Do(req)
			if err != nil {
				t.Fatal("http request", err)
			}

			if test.expectedStatus != 0 && test.expectedStatus != httpResp.StatusCode {
				t.Errorf("expected status code %d, got %d", test.expectedStatus, httpResp.StatusCode)
			}

			for k, v := range test.expectedHeaders {
				if diff := cmp.Diff([]string{v}, httpResp.Header[k]); diff != "" {
					t.Errorf("expected header %q %s", k, diff)
				}
			}

			body, err := io.ReadAll(httpResp.Body)
			if err != nil {
				t.Fatal(err)
			}

			if test.ignoreResponseBody {
				return
			}

			if test.expectedResponse == nil {
				if len(body) != 0 {
					t.Error("expected no body")
				}

				return
			}

			var errorResp httprouter.ErrorResponse
			if err := json.Unmarshal(body, &errorResp); err != nil {
				t.Error(err)
				return
			}

			if diff := cmp.Diff(test.expectedResponse, &errorResp); diff != "" {
				t.Error("expected response", diff)
			}
		})
	}
}

func TestRouteData(t *testing.T) {
	router := httprouter.New(httprouter.WithVerbose(true))

	observer := routeDataObserver{}
	handler := func(w http.ResponseWriter, req *http.Request) error {
		data := httprouter.GetRouteData(req.Context())
		observer.add(data)

		return nil
	}

	tests := []struct {
		name           string
		method         string
		route          string
		path           string
		expectedParams map[string]string
	}{
		{
			name:   "static route",
			method: http.MethodGet,
			path:   "/",
			route:  "/",
		},
		{
			name:   "multiple methods",
			method: http.MethodPost,
			path:   "/",
			route:  "/",
		},
		{
			name:   "path with one parameter",
			method: http.MethodGet,
			path:   "/user/1",
			route:  "/user/:id",
			expectedParams: map[string]string{
				"id": "1",
			},
		},
		{
			name:   "static path override",
			method: http.MethodGet,
			path:   "/user/special",
			route:  "/user/special",
		},
		{
			name:   "path with multiple parameters",
			method: http.MethodGet,
			path:   "/user/1/address/2",
			route:  "/user/:user_id/address/:address_id",
			expectedParams: map[string]string{
				"user_id":    "1",
				"address_id": "2",
			},
		},
		{
			name:   "partial path override",
			method: http.MethodGet,
			path:   "/user/special/address/1",
			route:  "/user/special/address/:address_id",
			expectedParams: map[string]string{
				"address_id": "1",
			},
		},
		{
			name:   "path with wildcard",
			method: http.MethodGet,
			path:   "/static/foo.txt",
			route:  "/static/*",
			expectedParams: map[string]string{
				"*": "foo.txt",
			},
		},
	}

	for _, test := range tests {
		router.Handler(test.method, test.route, handler)
	}

	// t.Log(router.DumpTree())

	server := httptest.NewServer(router)

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, server.URL+test.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := server.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if http.StatusOK != resp.StatusCode {
				t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
			}

			entries := observer.takeAll()
			if len(entries) != 1 {
				t.Fatalf("expected single route data entry, got %d", len(entries))
			}

			entry := entries[0]
			expectedEntry := httprouter.RouteData{
				Route:  test.route,
				Params: test.expectedParams,
			}

			if diff := cmp.Diff(expectedEntry, entry); diff != "" {
				t.Error("unexpected entry", diff)
			}
		})
	}

}

type routeDataObserver struct {
	entries []httprouter.RouteData
}

func (o *routeDataObserver) add(entry httprouter.RouteData) {
	o.entries = append(o.entries, entry)
}

func (o *routeDataObserver) takeAll() []httprouter.RouteData {
	dst := make([]httprouter.RouteData, len(o.entries))
	copy(dst, o.entries)
	o.entries = o.entries[:0]

	return dst
}
