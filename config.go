package httprouter

import (
	"net/http"
)

type Opt func(c *config)

type LogRoundtrip func(rw ResponseWriter, req *http.Request)

func NoopHandler(http.ResponseWriter, *http.Request) error {
	return nil
}

func WithVerbose(verbose bool) Opt {
	return func(c *config) {
		c.verbose = verbose
	}
}

func WithRedirectTrailingSlash(redirectTrailingSlash bool) Opt {
	return func(c *config) {
		c.redirectTrailingSlash = redirectTrailingSlash
	}
}

func WithLogRoundtrip(logRoundtrip LogRoundtrip) Opt {
	return func(c *config) {
		c.logRoundtrip = logRoundtrip
	}
}

func WithHandleOptions(handleOptions bool) Opt {
	return func(c *config) {
		c.handleOptions = handleOptions
	}
}

func WithErrorHandler(handler ErrorHandler) Opt {
	if handler == nil {
		panic("ErrorHandler cannot be nil")
	}

	return func(c *config) {
		c.errorHandler = handler
	}
}

func WithPanicHandler(handler PanicHandler) Opt {
	return func(c *config) {
		c.panicHandler = handler
	}
}

func WithOptionsHandler(handler HandlerFunc) Opt {
	return func(c *config) {
		c.optionsHandler = handler
	}
}

func WithMiddleware(middleware ...Middleware) Opt {
	return func(c *config) {
		c.middleware = append(c.middleware, middleware...)
	}
}

type config struct {
	verbose               bool
	handleOptions         bool
	redirectTrailingSlash bool
	logRoundtrip          LogRoundtrip
	errorHandler          ErrorHandler
	panicHandler          PanicHandler
	optionsHandler        HandlerFunc
	middleware            []Middleware
}

var defaultConfig = config{
	handleOptions:         true,
	redirectTrailingSlash: true,
	errorHandler:          DefaultErrorHandler,
	panicHandler:          DefaultPanicHandler,
	optionsHandler:        NoopHandler,
}
