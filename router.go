package httprouter

import (
	"net/http"

	"github.com/goes-funky/zapdriver"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

var ParamsFromContext = httprouter.ParamsFromContext

type HandlerFunc func(w http.ResponseWriter, req *http.Request) error

type Middleware func(handler HandlerFunc) HandlerFunc

type ErrorHandlerFunc func(w http.ResponseWriter, req *http.Request, err error)

type PanicHandlerFunc func(w http.ResponseWriter, req *http.Request, pv interface{})

type Router struct {
	logger       *zap.Logger
	config       config
	delegate     *httprouter.Router
	errorHandler ErrorHandlerFunc
}

func New(logger *zap.Logger, opts ...Opt) *Router {
	config := defaultConfig
	for _, opt := range opts {
		opt(&config)
	}

	errorHandler := config.errorHandler(logger, config.verbose)

	delegate := httprouter.New()
	delegate.HandleMethodNotAllowed = true
	delegate.NotFound = adaptHandler(logger, &config, errorHandler, func(http.ResponseWriter, *http.Request) error {
		return NewError(http.StatusNotFound)
	})
	delegate.MethodNotAllowed = adaptHandler(logger, &config, errorHandler, func(http.ResponseWriter, *http.Request) error {
		return NewError(http.StatusMethodNotAllowed)
	})
	delegate.PanicHandler = config.panicHandler(logger, config.verbose)

	if config.globalOptions != nil {
		delegate.HandleOPTIONS = true
		delegate.GlobalOPTIONS = adaptHandler(logger, &config, errorHandler, config.globalOptions)
	}

	return &Router{
		logger:       logger,
		config:       config,
		delegate:     delegate,
		errorHandler: errorHandler,
	}
}

func (r *Router) Handler(method, path string, handler HandlerFunc, middleware ...Middleware) {
	middleware = append(r.config.middleware, middleware...)

	for _, mw := range middleware {
		handler = mw(handler)
	}

	r.delegate.HandlerFunc(method, path, adaptHandler(r.logger, &r.config, r.errorHandler, handler))
}

func (r *Router) RawHandler(method, path string, handler http.Handler, middleware ...Middleware) {
	h := func(rw http.ResponseWriter, r *http.Request) error {
		handler.ServeHTTP(rw, r)
		return nil
	}

	r.Handler(method, path, h, middleware...)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.delegate.ServeHTTP(w, req)
}

func adaptHandler(logger *zap.Logger, config *config, errorHandler ErrorHandlerFunc, handler HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rw := NewResponseWriter(w)
		err := handler(rw, req)
		if err != nil {
			errorHandler(rw, req, err)
		}

		if config.logHTTP {
			payload := zapdriver.NewHTTP(req)
			payload.Status = rw.StatusCode()
			payload.ResponseSize = rw.Size()
			payload.Latency = rw.Latency()

			logger.WithOptions(zap.WithCaller(false)).Info("http request", zapdriver.HTTP(payload))
		}
	})
}
