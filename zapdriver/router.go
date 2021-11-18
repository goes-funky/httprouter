package zapdriver

import (
	"net/http"

	"github.com/goes-funky/httprouter"
	"github.com/goes-funky/zapdriver"

	"go.uber.org/zap"
)

func NewRouter(dev bool, opts ...httprouter.Opt) (*httprouter.Router, error) {
	var config zap.Config
	switch {
	case dev:
		config = zapdriver.NewDevelopmentConfig()
	default:
		config = zapdriver.NewProductionConfig()
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	opts = append([]httprouter.Opt{httprouter.WithLogRoundtrip(LogRoundtrip)}, opts...)

	return httprouter.New(logger, opts...), nil
}

func LogRoundtrip(logger *zap.Logger, rw *httprouter.ResponseWriter, req *http.Request) {
	payload := zapdriver.NewHTTP(req)
	payload.Status = rw.StatusCode()
	payload.ResponseSize = rw.Size()
	payload.Latency = rw.Latency()

	logger.WithOptions(zap.WithCaller(false)).Info("roundtrip", zapdriver.HTTP(payload))
}
