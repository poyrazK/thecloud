// Package main provides the cloud CLI entrypoint.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.4.0"

var rootCmd = &cobra.Command{
	Use:   "cloud",
	Short: "The Cloud Cloud CLI",
	Long:  `A local-first cloud simulator CLI for learning and development.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of cloud CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("cloud v%s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(instanceCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(storageCmd)
	rootCmd.AddCommand(vpcCmd)
	rootCmd.AddCommand(lbCmd)
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(secretsCmd)
	rootCmd.AddCommand(functionCmd)
	rootCmd.AddCommand(cacheCmd)
	rootCmd.AddCommand(autoscalingCmd)
	rootCmd.AddCommand(containerCmd)
	rootCmd.AddCommand(cronCmd)
	rootCmd.AddCommand(eventsCmd)
	rootCmd.AddCommand(gatewayCmd)
	rootCmd.AddCommand(notifyCmd)
	rootCmd.AddCommand(queueCmd)
	rootCmd.AddCommand(snapshotCmd)
	rootCmd.AddCommand(volumeCmd)
	rootCmd.AddCommand(iacCmd)
	rootCmd.AddCommand(rolesCmd)
	rootCmd.AddCommand(subnetCmd)
	rootCmd.AddCommand(kubernetesCmd)
	rootCmd.AddCommand(dnsCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
