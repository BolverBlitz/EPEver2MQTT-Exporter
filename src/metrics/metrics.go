package metrics

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricMap = make(map[string]*prometheus.GaugeVec)
	mutex     sync.Mutex
	namespace = getNamespace()

	customRegistry   = prometheus.NewRegistry()
	regAwarePromauto = promauto.With(customRegistry)
)

func GetCustomRegistry() prometheus.Gatherer {
	return customRegistry
}

func getNamespace() string {
	ns := os.Getenv("PROM_NAMESPACE")
	if ns == "" {
		ns = "default" // fallback if not set
	}
	return ns
}

// Init registers metrics from the config list
func Init(metricPaths []string) {
	for _, path := range metricPaths {
		createMetric(path)
	}
}

// sanitizeName ensures the metric name component is valid for Prometheus.
func sanitizeName(input string) string {
	name := strings.ReplaceAll(input, ".", "_")
	name = strings.ReplaceAll(name, "-", "_") // Replace hyphens as well
	return strings.ToLower(name)
}

func createMetric(path string) {
	mutex.Lock()
	defer mutex.Unlock()

	// Store the original path for use in the 'Help' string for clarity.
	originalPathForHelp := path

	if _, ok := metricMap[originalPathForHelp]; ok {
		return
	}

	var name, subsystem string
	parts := strings.Split(path, ".")

	if len(parts) > 1 {
		subsystem = sanitizeName(parts[0])
		name = sanitizeName(strings.Join(parts[1:], "_"))
	} else {
		subsystem = "general"
		name = sanitizeName(path)
	}

	// Ensure the metric name is not empty after sanitization.
	if name == "" {
		name = "unspecified_metric" // Provide a fallback name.
	}

	// Use the registry-aware promauto factory for creating the metric.
	metric := regAwarePromauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      fmt.Sprintf("Metric for %s", originalPathForHelp),
	}, []string{"id"})

	metricMap[originalPathForHelp] = metric
}

// UpdateMetric sets the value for a registered metric
func UpdateMetric(path string, value float64, id string) {
	mutex.Lock()
	defer mutex.Unlock()

	if metric, ok := metricMap[path]; ok {
		metric.WithLabelValues(id).Set(value)
	}
}
