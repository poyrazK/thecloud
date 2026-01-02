package platform

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	WSConnectionsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "mini_aws_ws_connections_active",
		Help: "The total number of active WebSocket connections",
	})
)
