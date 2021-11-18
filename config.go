package httprouter

import (
	"go.uber.org/zap"
)

type Opt func(c *config)

type ErrorHandler func(logger *zap.Logger, verbose bool) ErrorHandlerFunc

type PanicHandler func(logger *zap.Logger, verbose bool) PanicHandlerFunc

func WithVerbose(verbose bool) Opt {
	return func(c *config) {
		c.verbose = verbose
	}
}

func WithLogHTTP(logHTTP bool) Opt {
	return func(c *config) {
		c.logHTTP = logHTTP
	}
}

func WithErrorHandler(handler ErrorHandler) Opt {
	return func(c *config) {
		c.errorHandler = handler
	}
}

func WithPanicHandler(handler PanicHandler) Opt {
	return func(c *config) {
		c.panicHandler = handler
	}
}

func WithMiddleware(middleware ...Middleware) Opt {
	return func(c *config) {
		c.middleware = append(c.middleware, middleware...)
	}
}

func WithGlobalOptions(handler HandlerFunc) Opt {
	return func(c *config) {
		c.globalOptions = handler
	}
}

type config struct {
	verbose       bool
	logHTTP       bool
	errorHandler  ErrorHandler
	panicHandler  PanicHandler
	middleware    []Middleware
	globalOptions HandlerFunc
}

var defaultConfig = config{
	logHTTP:      true,
	errorHandler: DefaultErrorHandler,
	panicHandler: DefaultPanicHandler,
}
