package collector

func uptimeMetric(seconds float64) []Metric {
	return []Metric{{
		Name:  "gosysmon_system_uptime_seconds",
		Help:  "System uptime in seconds.",
		Type:  Gauge,
		Value: seconds,
	}}
}
