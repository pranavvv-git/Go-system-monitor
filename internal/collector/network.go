package collector

func networkMetrics(name string, rxBytes, txBytes float64) []Metric {
	labels := map[string]string{"interface": name}
	return []Metric{
		{
			Name:   "gosysmon_network_rx_bytes_total",
			Help:   "Total bytes received by the network interface.",
			Type:   Counter,
			Labels: labels,
			Value:  rxBytes,
		},
		{
			Name:   "gosysmon_network_tx_bytes_total",
			Help:   "Total bytes transmitted by the network interface.",
			Type:   Counter,
			Labels: labels,
			Value:  txBytes,
		},
	}
}
