package httpresolve

import (
	"errors"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"os"
	"strconv"
)

func init() { plugin.Register("httpresolve", setup) }

func setup(c *caddy.Controller) error {
	httpResolve, err := newHttpResolve(c)
	if err != nil {
		return plugin.Error("httpresolve", err)
	}
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		httpResolve.Next = next
		return httpResolve
	})

	return nil
}

func newHttpResolve(c *caddy.Controller) (*HttpResolve, error) {
	httpResolve := &HttpResolve{}
	for c.Next() {
		for c.NextBlock() {
			switch c.Val() {
			case "endpoint":
				httpResolve.Endpoint = os.Getenv(c.RemainingArgs()[0])
			case "secretId":
				httpResolve.SecretId = os.Getenv(c.RemainingArgs()[0])
			case "secretKey":
				httpResolve.SecretKey = os.Getenv(c.RemainingArgs()[0])
			case "ttl":
				ttlString := os.Getenv(c.RemainingArgs()[0])
				ttl64, _ := strconv.ParseFloat(ttlString, 64)
				httpResolve.Ttl = uint32(ttl64)
			}
		}

	}

	if httpResolve.Endpoint == "" || httpResolve.SecretId == "" || httpResolve.SecretKey == "" {
		return nil, errors.New("params[endpoint, secretId, secretKey] cannot be empty")
	}

	return httpResolve, nil
}
