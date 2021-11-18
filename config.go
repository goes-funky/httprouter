package httprouter

import (
	"net/http"

	"go.uber.org/zap"
)

type Opt func(c *config)

type LogRoundtrip func(logger *zap.Logger, rw *ResponseWriter, req *http.Request)

type ErrorHandler func(logger *zap.Logger, verbose bool) ErrorHandlerFunc

type PanicHandler func(logger *zap.Logger, verbose bool) PanicHandlerFunc

func DefaultLogRoundtrip(logger *zap.Logger, rw *ResponseWriter, req *http.Request) {
	logger.Info("roundtrip",
		zap.String("method", req.Method),
		zap.String("path", req.URL.Path),
		zap.Int("status", rw.StatusCode()),
		zap.Duration("elapsed", rw.Latency()),
	)
}

func WithVerbose(verbose bool) Opt {
	return func(c *config) {
		c.verbose = verbose
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
	handleOptions bool
	logRoundtrip  LogRoundtrip
	errorHandler  ErrorHandler
	panicHandler  PanicHandler
	middleware    []Middleware
	globalOptions HandlerFunc
}

var defaultConfig = config{
	handleOptions: true,
	logRoundtrip:  DefaultLogRoundtrip,
	errorHandler:  DefaultErrorHandler,
	panicHandler:  DefaultPanicHandler,
}
