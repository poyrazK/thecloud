package main

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/util/wait"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/cloud-provider/app"
	"k8s.io/cloud-provider/app/config"
	"k8s.io/cloud-provider/options"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	_ "k8s.io/component-base/metrics/prometheus/clientgo" // load all the prometheus client-go plugins
	_ "k8s.io/component-base/metrics/prometheus/version"  // for version metric registration

	"github.com/spf13/pflag"

	_ "github.com/poyrazk/thecloud/internal/ccm" // Register the cloud provider
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	logs.InitLogs()
	defer logs.FlushLogs()

	ccmOptions, err := options.NewCloudControllerManagerOptions()
	if err != nil {
		return fmt.Errorf("unable to initialize command options: %w", err)
	}

	command := app.NewCloudControllerManagerCommand(
		ccmOptions,
		cloudInitializer,
		app.DefaultInitFuncConstructors,
		map[string]string{},
		cliflag.NamedFlagSets{},
		wait.NeverStop,
	)

	// Explicitly add Go flags to the command (for klog)
	pflag.CommandLine.AddGoFlagSet(nil)

	return command.Execute()
}

func cloudInitializer(config *config.CompletedConfig) cloudprovider.Interface {
	cloudConfig := config.ComponentConfig.KubeCloudShared.CloudProvider

	// Initialize the cloud provider
	cloud, err := cloudprovider.InitCloudProvider(cloudConfig.Name, cloudConfig.CloudConfigFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cloud provider could not be initialized: %v\n", err)
		os.Exit(1)
	}
	if cloud == nil {
		fmt.Fprintf(os.Stderr, "Cloud provider is nil\n")
		os.Exit(1)
	}

	if !cloud.HasClusterID() {
		if config.ComponentConfig.KubeCloudShared.AllowUntaggedCloud {
			fmt.Println("warning: cluster ID is not set. Many cloud providers require this.")
		} else {
			fmt.Fprintf(os.Stderr, "error: cluster ID is required but not set\n")
			os.Exit(1)
		}
	}

	return cloud
}
