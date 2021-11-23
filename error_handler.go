package httprouter

import (
	"encoding/json"
	"net/http"
	"strings"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Debug   string `json:"debug,omitempty"`
}

func DefaultErrorHandler(w http.ResponseWriter, req *http.Request, verbose bool, err error) {
	httpErr := AsError(err)
	cause := httpErr.Cause

	// do not write JSON response on http methods that do not return body
	if req.Method == http.MethodHead || req.Method == http.MethodPut || req.Method == http.MethodTrace {
		w.WriteHeader(httpErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpErr.Status)

	var debug string
	if verbose && cause != nil {
		debug = cause.Error()
	}

	resp := ErrorResponse{
		Message: httpErr.Message,
		Debug:   debug,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func DefaultPanicHandler(w http.ResponseWriter, req *http.Request, verbose bool, pv interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	resp := ErrorResponse{
		Message: http.StatusText(http.StatusInternalServerError),
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func DefaultMethodNotAllowed(rw http.ResponseWriter, req *http.Request, verbose bool, methods []string) {
	rw.Header().Set("Allow", strings.Join(methods, ", "))
	DefaultErrorHandler(rw, req, verbose, NewError(http.StatusMethodNotAllowed))
}
