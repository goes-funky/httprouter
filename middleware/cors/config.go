package cors

import (
	"strconv"
	"strings"
	"time"
)

type Opt func(c *config)

func WithOrigin(origin string) Opt {
	return func(c *config) {
		c.origin = origin
	}
}

func WithRequestHeaders(headers []string) Opt {
	return func(c *config) {
		c.requestHeaders = strings.Join(headers, ", ")
	}
}

func WithAllowCredentials(allowCredentials bool) Opt {
	return func(c *config) {
		c.allowCredentials = allowCredentials
	}
}

func MaxAge(maxAge time.Duration) Opt {
	return func(c *config) {
		c.maxAge = strconv.FormatInt(int64(maxAge), 10)
	}
}

type config struct {
	origin           string
	requestHeaders   string
	allowCredentials bool
	maxAge           string
}

func (c config) originWildcard() bool {
	return c.origin == "*"
}

func (c config) requestHeadersWildcard() bool {
	return c.requestHeaders == "*"
}

var defaultConfig = config{
	origin:           "*",
	requestHeaders:   "*",
	allowCredentials: true,
}
