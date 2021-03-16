package httpresolve

import (
	"github.com/coredns/coredns/plugin"
	"github.com/prometheus/client_golang/prometheus"
)

var requestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: plugin.Namespace,
	Subsystem: "httpresolve",
	Name:      "request_count_total",
	Help:      "Counter of requests made.",
}, []string{"server"})
