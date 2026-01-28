// Package platform provides shared infrastructure utilities.
package platform

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// StorageOperations counts total storage operations by type and status
	StorageOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "storage_operations_total",
			Help: "Total storage operations",
		},
		[]string{"operation", "bucket", "status"},
	)

	// StorageLatency measures latency of storage operations
	StorageLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "storage_operation_duration_seconds",
			Help:    "Storage operation latency",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "bucket"},
	)

	// StorageBytesTransferred counts total bytes uploaded/downloaded
	StorageBytesTransferred = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "storage_bytes_transferred_total",
			Help: "Total bytes transferred to/from storage",
		},
		[]string{"direction"}, // "upload", "download"
	)

	// StorageBucketObjects counts total objects per bucket (gauge)
	StorageBucketObjects = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "storage_bucket_objects_total",
			Help: "Total number of objects per bucket",
		},
		[]string{"bucket"},
	)

	// StorageBucketBytes counts total bytes stored per bucket (gauge)
	StorageBucketBytes = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "storage_bucket_bytes_total",
			Help: "Total bytes stored per bucket",
		},
		[]string{"bucket"},
	)
)
