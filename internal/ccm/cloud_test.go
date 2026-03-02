package ccm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudprovider "k8s.io/cloud-provider"
)

func TestProviderRegistration(t *testing.T) {
	t.Setenv("CLOUD_API_KEY", "dummy")
	cloud, err := cloudprovider.GetCloudProvider(ProviderName, nil)
	require.NoError(t, err)
	assert.NotNil(t, cloud)
	assert.Equal(t, ProviderName, cloud.ProviderName())

	tests := []struct {
		name      string
		fn        func() (interface{}, bool)
		supported bool
	}{
		{
			name: "LoadBalancer",
			fn: func() (interface{}, bool) {
				lb, ok := cloud.LoadBalancer()
				return lb, ok
			},
			supported: true,
		},
		{
			name: "Instances",
			fn: func() (interface{}, bool) {
				inst, ok := cloud.Instances()
				return inst, ok
			},
			supported: false,
		},
		{
			name: "InstancesV2",
			fn: func() (interface{}, bool) {
				instV2, ok := cloud.InstancesV2()
				return instV2, ok
			},
			supported: true,
		},
		{
			name: "Zones",
			fn: func() (interface{}, bool) {
				zones, ok := cloud.Zones()
				return zones, ok
			},
			supported: false,
		},
		{
			name: "Clusters",
			fn: func() (interface{}, bool) {
				clusters, ok := cloud.Clusters()
				return clusters, ok
			},
			supported: false,
		},
		{
			name: "Routes",
			fn: func() (interface{}, bool) {
				routes, ok := cloud.Routes()
				return routes, ok
			},
			supported: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, ok := tt.fn()
			assert.Equal(t, tt.supported, ok)
			if tt.supported {
				assert.NotNil(t, obj)
			} else {
				assert.Nil(t, obj)
			}
		})
	}

	assert.False(t, cloud.HasClusterID())

	// Initialize shouldn't crash
	cloud.Initialize(nil, nil)
}
