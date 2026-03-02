package ccm

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cloudprovider "k8s.io/cloud-provider"
)

func TestProviderRegistration(t *testing.T) {
	os.Setenv("CLOUD_API_KEY", "dummy")
	cloud, err := cloudprovider.GetCloudProvider(ProviderName, nil)
	require.NoError(t, err)
	assert.NotNil(t, cloud)
	assert.Equal(t, ProviderName, cloud.ProviderName())
	
	// Test other interface methods
	lb, supported := cloud.LoadBalancer()
	assert.True(t, supported)
	assert.NotNil(t, lb)

	inst, supported := cloud.Instances()
	assert.False(t, supported)
	assert.Nil(t, inst)

	instV2, supported := cloud.InstancesV2()
	assert.True(t, supported)
	assert.NotNil(t, instV2)

	zones, supported := cloud.Zones()
	assert.False(t, supported)
	assert.Nil(t, zones)

	clusters, supported := cloud.Clusters()
	assert.False(t, supported)
	assert.Nil(t, clusters)

	routes, supported := cloud.Routes()
	assert.False(t, supported)
	assert.Nil(t, routes)

	assert.False(t, cloud.HasClusterID())
	
	// Initialize shouldn't crash
	cloud.Initialize(nil, nil)
}
