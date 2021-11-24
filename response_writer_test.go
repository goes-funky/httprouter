package httprouter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goes-funky/httprouter"
)

func TestResponseWriterFlusherPusher(t *testing.T) {
	router := httprouter.New()
	router.Handler(http.MethodGet, "/", func(rw http.ResponseWriter, req *http.Request) error {
		if _, ok := rw.(http.Flusher); !ok {
			t.Error("ResponseWriter does not implement http.Flusher interface")
		}

		if _, ok := rw.(http.Pusher); !ok {
			t.Error("ResponseWriter does not implement http.Pusher interface")
		}

		status := rw.(httprouter.ResponseWriter).StatusCode()

		if http.StatusOK != status {
			t.Errorf("expected status %d, got %d", http.StatusOK, status)
		}

		return nil
	})

	server := httptest.NewServer(router)
	resp, err := server.Client().Get(server.URL + "/")
	if err != nil {
		t.Error("expected GET / to succeed")
	}

	if http.StatusOK != resp.StatusCode {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
