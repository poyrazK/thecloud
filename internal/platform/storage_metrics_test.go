package platform

import (
	"reflect"
	"sort"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_model/go"
)

func TestStorageMetricsRegistered(t *testing.T) {
	StorageOperations.WithLabelValues("put", "bucket-a", "ok").Inc()
	StorageLatency.WithLabelValues("get", "bucket-a").Observe(0.01)
	StorageBytesTransferred.WithLabelValues("upload").Add(1)
	StorageBucketObjects.WithLabelValues("bucket-a").Set(1)
	StorageBucketBytes.WithLabelValues("bucket-a").Set(1)

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	expected := map[string][]string{
		"storage_operations_total":           {"operation", "bucket", "status"},
		"storage_operation_duration_seconds": {"operation", "bucket"},
		"storage_bytes_transferred_total":    {"direction"},
		"storage_bucket_objects_total":       {"bucket"},
		"storage_bucket_bytes_total":         {"bucket"},
	}

	for name, labels := range expected {
		family := findMetricFamily(families, name)
		actualLabels := labelNames(family)
		sortedExpected := append([]string(nil), labels...)
		sort.Strings(sortedExpected)
		if !reflect.DeepEqual(actualLabels, sortedExpected) {
			t.Fatalf("metric %s labels mismatch: got %v want %v", name, actualLabels, labels)
		}
	}
}

func findMetricFamily(families []*io_prometheus_client.MetricFamily, name string) *io_prometheus_client.MetricFamily {
	for _, family := range families {
		if family.GetName() == name {
			return family
		}
	}
	return nil
}

func labelNames(family *io_prometheus_client.MetricFamily) []string {
	if family == nil || len(family.Metric) == 0 || len(family.Metric[0].Label) == 0 {
		return nil
	}

	labels := make([]string, 0, len(family.Metric[0].Label))
	for _, label := range family.Metric[0].Label {
		labels = append(labels, label.GetName())
	}
	sort.Strings(labels)
	return labels
}
