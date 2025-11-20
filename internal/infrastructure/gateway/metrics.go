package gateway

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type metricKey struct {
	Route  string
	Status int
}

type latencyStat struct {
	Total time.Duration
	Count int64
}

type gatewayMetrics struct {
	mu        sync.Mutex
	requests  map[metricKey]int64
	latencies map[string]latencyStat
}

func newGatewayMetrics() *gatewayMetrics {
	return &gatewayMetrics{
		requests:  map[metricKey]int64{},
		latencies: map[string]latencyStat{},
	}
}

func (m *gatewayMetrics) Observe(route string, status int, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := metricKey{Route: route, Status: status}
	m.requests[key]++

	stat := m.latencies[route]
	stat.Total += latency
	stat.Count++
	m.latencies[route] = stat
}

// Format renders metrics in a Prometheus-friendly text format.
func (m *gatewayMetrics) Format() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	lines := []string{}

	keys := make([]metricKey, 0, len(m.requests))
	for k := range m.requests {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Route == keys[j].Route {
			return keys[i].Status < keys[j].Status
		}
		return keys[i].Route < keys[j].Route
	})

	for _, k := range keys {
		lines = append(lines, fmt.Sprintf(`gateway_requests_total{route="%s",status="%d"} %d`, k.Route, k.Status, m.requests[k]))
	}

	routes := make([]string, 0, len(m.latencies))
	for route := range m.latencies {
		routes = append(routes, route)
	}
	sort.Strings(routes)
	for _, route := range routes {
		stat := m.latencies[route]
		lines = append(lines, fmt.Sprintf(`gateway_request_latency_ms_sum{route="%s"} %.0f`, route, stat.Total.Seconds()*1000))
		lines = append(lines, fmt.Sprintf(`gateway_request_latency_ms_count{route="%s"} %d`, route, stat.Count))
	}

	return lines
}
