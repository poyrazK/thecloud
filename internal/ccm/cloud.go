package ccm

import (
	"fmt"
	"io"
	"os"

	"github.com/poyrazk/thecloud/pkg/sdk"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"
)

const ProviderName = "thecloud"

type CloudProvider struct {
	client *sdk.Client
	lb     cloudprovider.LoadBalancer
	instV2 cloudprovider.InstancesV2
}

func init() {
	cloudprovider.RegisterCloudProvider(ProviderName, func(config io.Reader) (cloudprovider.Interface, error) {
		apiURL := os.Getenv("CLOUD_API_URL")
		apiKey := os.Getenv("CLOUD_API_KEY")

		if apiURL == "" {
			apiURL = "http://thecloud-api.kube-system.svc.cluster.local:8080"
		}
		if apiKey == "" {
			return nil, fmt.Errorf("CLOUD_API_KEY is required for the cloud controller manager")
		}

		client := sdk.NewClient(apiURL, apiKey)

		return &CloudProvider{
			client: client,
			lb:     newLoadBalancer(client),
			instV2: newInstancesV2(client),
		}, nil
	})
}

func (c *CloudProvider) Initialize(clientBuilder cloudprovider.ControllerClientBuilder, stop <-chan struct{}) {
	klog.Infof("Initializing The Cloud CCM")
}

func (c *CloudProvider) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	return c.lb, true
}

func (c *CloudProvider) Instances() (cloudprovider.Instances, bool) {
	return nil, false
}

func (c *CloudProvider) InstancesV2() (cloudprovider.InstancesV2, bool) {
	return c.instV2, true
}

func (c *CloudProvider) Zones() (cloudprovider.Zones, bool) {
	return nil, false
}

func (c *CloudProvider) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

func (c *CloudProvider) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

func (c *CloudProvider) ProviderName() string {
	return ProviderName
}

func (c *CloudProvider) HasClusterID() bool {
	return false
}
