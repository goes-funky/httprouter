package zapdriver

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/goes-funky/httprouter"
	"github.com/goes-funky/zapdriver"
)

var NewDevelopmentConfig = zapdriver.NewDevelopmentConfig
var NewProductionConfig = zapdriver.NewProductionConfig

func LogRoundtrip(logger *zap.Logger) httprouter.LogRoundtrip {
	return func(rw httprouter.ResponseWriter, req *http.Request) {
		payload := zapdriver.NewHTTP(req)
		payload.Status = rw.StatusCode()
		payload.ResponseSize = rw.Size()
		payload.Latency = rw.Latency()

		logger.WithOptions(zap.WithCaller(false)).Info("roundtrip", zapdriver.HTTP(payload))
	}
}
