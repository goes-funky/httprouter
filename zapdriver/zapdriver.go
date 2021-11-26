package zapdriver

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/goes-funky/httprouter"
	"github.com/goes-funky/zapdriver"
)

var (
	NewDevelopmentConfig = zapdriver.NewDevelopmentConfig
	NewProductionConfig  = zapdriver.NewProductionConfig
)

func RouterOpts(logger *zap.Logger) []httprouter.Opt {
	return []httprouter.Opt{
		LogRoundtrip(logger),
		ErrorHandler(logger, nil),
	}
}

func LogRoundtrip(logger *zap.Logger) httprouter.Opt {
	return httprouter.WithLogRoundtrip(func(rw httprouter.ResponseWriter, req *http.Request) {
		payload := zapdriver.NewHTTP(req)
		payload.Status = rw.StatusCode()
		payload.ResponseSize = rw.Size()
		payload.Latency = rw.Latency()

		logger.WithOptions(zap.WithCaller(false)).Info("roundtrip", zapdriver.HTTP(payload))
	})
}

func ErrorHandler(logger *zap.Logger, next httprouter.ErrorHandler) httprouter.Opt {
	if next == nil {
		next = httprouter.DefaultErrorHandler
	}

	return httprouter.WithErrorHandler(func(w http.ResponseWriter, req *http.Request, verbose bool, err httprouter.Error) {
		if !err.Operational {
			fields := []zap.Field{
				zap.String("path", req.URL.Path),
				zap.Int("status", err.Status),
				zap.String("message", err.Message),
			}

			if err.Cause != nil {
				fields = append(fields, zap.Error(err.Cause))
			}

			logger.Info("http error", fields...)
		}

		next(w, req, verbose, err)
	})
}
