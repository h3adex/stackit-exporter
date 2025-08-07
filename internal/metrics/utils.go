package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

func normalize(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "-", "_"))
}

// SetOneHotStatus updates all binary status metrics so only the correct one is set to 1.
func SetOneHotStatus(metricMap map[string]*prometheus.GaugeVec, current string, labels prometheus.Labels) {
	for key, gaugeVec := range metricMap {
		if normalize(current) == normalize(key) {
			gaugeVec.With(labels).Set(1)
		} else {
			gaugeVec.With(labels).Set(0)
		}
	}
}
