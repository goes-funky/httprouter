package httprouter

import (
	"fmt"
	"net/http"
	"strings"
)

type HandlerFunc func(w http.ResponseWriter, req *http.Request) error

type Middleware func(handler HandlerFunc) HandlerFunc

type ErrorHandler func(w http.ResponseWriter, req *http.Request, verbose bool, err Error)

type PanicHandler func(rw http.ResponseWriter, req *http.Request, verbose bool, pv interface{})

type Router struct {
	config config
	root   *node
}

// LookupResult contains information about a route lookup, which is returned from Lookup and
// can be passed to ServeLookupResult if the request should be served.
type LookupResult struct {
	RouteData
	// StatusCode informs the caller about the result of the lookup.
	// This will generally be `http.StatusNotFound` or `http.StatusMethodNotAllowed` for an
	// error case. On a normal success, the statusCode will be `http.StatusOK`. A redirect code
	// will also be used in the case
	Status  int
	Handler HandlerFunc
	Methods []string // Only has a value when StatusCode is MethodNotAllowed.
}

func New(opts ...Opt) *Router {
	config := defaultConfig
	for _, opt := range opts {
		opt(&config)
	}

	return &Router{
		root:   &node{path: "/"},
		config: config,
	}
}

// Handler registers HandlerFunc at given method and path
func (r *Router) Handler(method, path string, handler HandlerFunc, middleware ...Middleware) {
	switch {
	case len(path) == 0:
		panic("Path must be non empty")
	case len(path) > 0 && path[0] != '/':
		panic(fmt.Sprintf("Path %q must start with slash", path))
	}

	middleware = append(r.config.middleware, middleware...)

	if len(middleware) > 0 {
		for i := len(middleware) - 1; i != 0; i-- {
			handler = middleware[i](handler)
		}
	}

	r.root.registerPath(method, path, handler, r.config.redirectTrailingSlash)
}

// HTTPHandler register http.Handler at given method and path
func (r *Router) HTTPHandler(method, path string, handler http.Handler, middleware ...Middleware) {
	h := func(rw http.ResponseWriter, r *http.Request) error {
		handler.ServeHTTP(rw, r)
		return nil
	}

	r.Handler(method, path, h, middleware...)
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	lr := r.Lookup(rw, req)
	r.ServeLookupResult(rw, req, lr)
}

func (r *Router) Lookup(rw http.ResponseWriter, req *http.Request) LookupResult {
	path := req.URL.Path

	trailingSlash := strings.HasSuffix(path, "/") && len(path) > 1
	if trailingSlash && r.config.redirectTrailingSlash {
		path = path[:len(path)-1]
	}

	n, handler, params := r.root.search(req.Method, path[1:])
	if n == nil {
		return LookupResult{
			Status: http.StatusNotFound,
		}
	}

	var paramsMap map[string]string
	if len(params) != 0 {
		if len(params) != len(n.leafWildcardNames) {
			panic(fmt.Sprintf("httprouter parameter list length mismatch: %v, %v",
				params, n.leafWildcardNames))
		}

		paramsMap = make(map[string]string)
		numParams := len(params)

		for index := 0; index < numParams; index++ {
			name := n.leafWildcardNames[numParams-index-1]
			if len(name) == 0 {
				name = "*"
			}

			paramsMap[name] = params[index]
		}
	}

	routeData := RouteData{
		Route:  n.route,
		Params: paramsMap,
	}

	if handler == nil {
		var methods []string
		for method := range n.leafHandler {
			methods = append(methods, method)
		}

		if _, ok := n.leafHandler[http.MethodOptions]; !ok && r.config.handleOptions {
			methods = append(methods, "OPTIONS")
		}

		if req.Method == "OPTIONS" && r.config.optionsHandler != nil {
			return LookupResult{
				Status:    http.StatusOK,
				RouteData: routeData,
				Handler:   r.config.optionsHandler,
				Methods:   methods,
			}
		}

		return LookupResult{
			Status:    http.StatusMethodNotAllowed,
			RouteData: routeData,
			Methods:   methods,
		}
	}

	if !n.isCatchAll && trailingSlash != n.addSlash && r.config.redirectTrailingSlash {
		var status int
		switch {
		case req.Method != http.MethodGet:
			status = http.StatusTemporaryRedirect
		default:
			status = http.StatusPermanentRedirect
		}

		if n.addSlash {
			path += "/"
		}

		return LookupResult{
			Status: status,
			Handler: func(rw http.ResponseWriter, req *http.Request) error {
				http.Redirect(rw, req, path, status)
				return nil
			},
			RouteData: routeData,
		}
	}

	return LookupResult{
		Status:    http.StatusOK,
		Handler:   handler,
		RouteData: routeData,
	}
}

func (r *Router) ServeLookupResult(rw http.ResponseWriter, req *http.Request, lr LookupResult) {
	w := NewResponseWriter(rw)
	ctx := WithRouteData(req.Context(), lr.RouteData)
	req = req.WithContext(ctx)

	if r.config.panicHandler != nil {
		defer func() {
			if pv := recover(); pv != nil {
				r.config.panicHandler(w, req, r.config.verbose, pv)
			}
		}()
	}

	if r.config.logRoundtrip != nil {
		defer r.config.logRoundtrip(w, req)
	}

	if lr.Handler == nil {
		r.config.errorHandler(w, req, r.config.verbose, NewError(lr.Status, Operational()))
		return
	}

	if len(lr.Methods) != 0 {
		w.Header().Set("Allow", strings.Join(lr.Methods, ", "))
	}

	err := lr.Handler(w, req)
	if err != nil {
		r.config.errorHandler(w, req, r.config.verbose, AsError(err))
	}
}

func (r *Router) DumpTree() string {
	return r.root.dumpTree("", "")
}
