package zapdriver

import (
	"net/http"

	"github.com/goes-funky/httprouter"
	"github.com/goes-funky/zapdriver"

	"go.uber.org/zap"
)

func NewConfig(verbose bool) zap.Config {
	if verbose {
		return zapdriver.NewDevelopmentConfig()
	}

	return zapdriver.NewProductionConfig()
}

func RouterOpts(logger *zap.Logger, verbose bool) []httprouter.Opt {
	return []httprouter.Opt{
		httprouter.WithLogRoundtrip(LogRoundtrip(logger)),
		httprouter.WithVerbose(verbose),
	}
}

func LogRoundtrip(logger *zap.Logger) httprouter.LogRoundtrip {
	return func(rw httprouter.ResponseWriter, req *http.Request) {
		payload := zapdriver.NewHTTP(req)
		payload.Status = rw.StatusCode()
		payload.ResponseSize = rw.Size()
		payload.Latency = rw.Latency()

		logger.WithOptions(zap.WithCaller(false)).Info("roundtrip", zapdriver.HTTP(payload))
	}
}
