// Package main provides the cloud CLI entrypoint.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/poyrazk/thecloud/pkg/sdk"
	"github.com/spf13/cobra"
)

var apiURL = "http://localhost:8080"
var outputJSON bool
var apiKey string

const (
	fmtErrorLog     = "Error: %v\n"
	fmtDetailRow    = "%-15s %v\n"
	demoPrompt      = "[WARN] No API Key found. Run 'cloud auth create-demo <name>' to get one."
	successInstance = "[SUCCESS] Instance launched successfully!\n"
	infoStop        = "[INFO] Instance stop initiated."
)

var instanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Manage compute instances",
}

func getClient() *sdk.Client {
	key := apiKey // 1. Flag
	if key == "" {
		key = os.Getenv("CLOUD_API_KEY") // 2. Env Var
	}
	if key == "" {
		key = loadConfig() // 3. Config File
	}

	if key == "" {
		fmt.Println(demoPrompt)
		os.Exit(1)
	}

	return sdk.NewClient(apiURL, key)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all instances",
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		instances, err := client.ListInstances()
		if err != nil {
			fmt.Printf(fmtErrorLog, err)
			return
		}

		if outputJSON {
			data, _ := json.MarshalIndent(instances, "", "  ")
			fmt.Println(string(data))
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ID", "NAME", "IMAGE", "STATUS", "ACCESS"})

		for _, inst := range instances {
			id := inst.ID
			if len(id) > 8 {
				id = id[:8]
			}

			access := formatAccessPorts(inst.Ports, inst.Status)

			_ = table.Append([]string{
				id,
				inst.Name,
				inst.Image,
				inst.Status,
				access,
			})
		}
		_ = table.Render()
	},
}

var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch a new instance",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		image, _ := cmd.Flags().GetString("image")
		ports, _ := cmd.Flags().GetString("port")
		instanceType, _ := cmd.Flags().GetString("type")
		vpc, _ := cmd.Flags().GetString("vpc")
		subnetID, _ := cmd.Flags().GetString("subnet")
		volumeStrs, _ := cmd.Flags().GetStringSlice("volume")

		// Parse volume strings like "vol-name:/path"
		var volumes []sdk.VolumeAttachmentInput
		for _, v := range volumeStrs {
			parts := strings.SplitN(v, ":", 2)
			if len(parts) == 2 {
				volumes = append(volumes, sdk.VolumeAttachmentInput{
					VolumeID:  parts[0],
					MountPath: parts[1],
				})
			}
		}

		client := getClient()
		inst, err := client.LaunchInstance(name, image, ports, instanceType, vpc, subnetID, volumes)
		if err != nil {
			fmt.Printf(fmtErrorLog, err)
			return
		}

		fmt.Print(successInstance)
		data, _ := json.MarshalIndent(inst, "", "  ")
		fmt.Println(string(data))
	},
}
var stopCmd = &cobra.Command{
	Use:   "stop [id]",
	Short: "Stop an instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		client := getClient()
		if err := client.StopInstance(id); err != nil {
			fmt.Printf(fmtErrorLog, err)
			return
		}

		fmt.Println(infoStop)
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs [id]",
	Short: "View instance logs",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		client := getClient()
		logs, err := client.GetInstanceLogs(id)
		if err != nil {
			fmt.Printf(fmtErrorLog, err)
			return
		}

		fmt.Print(logs)
	},
}

var showCmd = &cobra.Command{
	Use:   "show [id/name]",
	Short: "Show detailed instance information",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		client := getClient()
		inst, err := client.GetInstance(id)
		if err != nil {
			fmt.Printf(fmtErrorLog, err)
			return
		}

		fmt.Print("\nInstance Details\n")
		fmt.Println(strings.Repeat("-", 40))
		fmt.Printf(fmtDetailRow, "ID:", inst.ID)
		fmt.Printf(fmtDetailRow, "Name:", inst.Name)
		fmt.Printf(fmtDetailRow, "Status:", inst.Status)
		fmt.Printf(fmtDetailRow, "Image:", inst.Image)
		fmt.Printf(fmtDetailRow, "Ports:", inst.Ports)
		fmt.Printf(fmtDetailRow, "Created At:", inst.CreatedAt)
		fmt.Printf(fmtDetailRow, "Version:", inst.Version)
		fmt.Printf(fmtDetailRow, "Container ID:", inst.ContainerID)
		fmt.Println(strings.Repeat("-", 40))
		fmt.Println("")
	},
}

var rmCmd = &cobra.Command{
	Use:   "rm [id/name]",
	Short: "Remove an instance and its resources",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		client := getClient()
		if err := client.TerminateInstance(id); err != nil {
			fmt.Printf(fmtErrorLog, err)
			return
		}

	},
}

var statsCmd = &cobra.Command{
	Use:   "stats [id/name]",
	Short: "Show instance statistics (CPU/Mem)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		client := getClient()
		stats, err := client.GetInstanceStats(id)
		if err != nil {
			fmt.Printf(fmtErrorLog, err)
			return
		}

		if outputJSON {
			data, _ := json.MarshalIndent(stats, "", "  ")
			fmt.Println(string(data))
			return
		}

		fmt.Printf("\nStatistics for %s\n", id)
		fmt.Println(strings.Repeat("-", 40))
		fmt.Printf("CPU:    %.2f%%\n", stats.CPUPercentage)
		fmt.Printf("Memory: %.2f%% (%.2f MB / %.2f MB)\n",
			stats.MemoryPercentage,
			stats.MemoryUsageBytes/1024/1024,
			stats.MemoryLimitBytes/1024/1024)
	},
}

func init() {
	instanceCmd.AddCommand(listCmd)
	instanceCmd.AddCommand(launchCmd)
	instanceCmd.AddCommand(stopCmd)
	instanceCmd.AddCommand(logsCmd)
	instanceCmd.AddCommand(showCmd)
	instanceCmd.AddCommand(rmCmd)
	instanceCmd.AddCommand(statsCmd)

	launchCmd.Flags().StringP("name", "n", "", "Name of the instance (required)")
	launchCmd.Flags().StringP("image", "i", "alpine", "Image to use")
	launchCmd.Flags().StringP("port", "p", "", "Port mapping (host:container)")
	launchCmd.Flags().StringP("type", "t", "basic-2", "Instance type (e.g. basic-1, standard-1)")
	launchCmd.Flags().StringP("vpc", "v", "", "VPC ID or Name to attach to")
	launchCmd.Flags().StringP("subnet", "s", "", "Subnet ID or Name to attach to")
	launchCmd.Flags().StringSliceP("volume", "V", nil, "Volume attachment (vol-name:/path)")
	_ = launchCmd.MarkFlagRequired("name")

	rootCmd.PersistentFlags().BoolVarP(&outputJSON, "json", "j", false, "Output in JSON format")
	rootCmd.PersistentFlags().StringVarP(&apiKey, "api-key", "k", "", "API key for authentication")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "http://localhost:8080", "URL of the API server")
}

func formatAccessPorts(ports string, status string) string {
	if ports == "" || status != "RUNNING" {
		return "-"
	}

	pList := strings.Split(ports, ",")
	var mappings []string
	for _, mapping := range pList {
		parts := strings.Split(mapping, ":")
		if len(parts) == 2 {
			mappings = append(mappings, fmt.Sprintf("localhost:%s->%s", parts[0], parts[1]))
		}
	}

	if len(mappings) == 0 {
		return ""
	}

	return strings.Join(mappings, ", ")
}
