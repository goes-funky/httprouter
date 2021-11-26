package httprouter

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Debug   string `json:"debug,omitempty"`
}

func DefaultErrorHandler(w http.ResponseWriter, req *http.Request, verbose bool, err Error) {
	// do not write JSON response on http methods that do not return body
	if req.Method == http.MethodHead || req.Method == http.MethodPut || req.Method == http.MethodTrace {
		w.WriteHeader(err.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)

	var debug string
	if verbose && err.Cause != nil {
		debug = err.Cause.Error()
	}

	resp := ErrorResponse{
		Message: err.Message,
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
