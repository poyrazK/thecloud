// Package main provides the cloud CLI entrypoint.
package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var containerCmd = &cobra.Command{
	Use:   "container",
	Short: "Manage CloudContainers (Deployments)",
}

var createDeploymentCmd = &cobra.Command{
	Use:   "deploy [name] [image]",
	Short: "Create a new container deployment",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		replicas, _ := cmd.Flags().GetInt("replicas")
		ports, _ := cmd.Flags().GetString("ports")

		client := getClient()
		dep, err := client.CreateDeployment(args[0], args[1], replicas, ports)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("[SUCCESS] Deployment created: %s (Status: %s)\n", dep.Name, dep.Status)
	},
}

var listDeploymentsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all deployments",
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		deps, err := client.ListDeployments()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ID", "NAME", "IMAGE", "REPLICAS", "CURRENT", "STATUS"})
		for _, d := range deps {
			table.Append([]string{
				d.ID,
				d.Name,
				d.Image,
				fmt.Sprintf("%d", d.Replicas),
				fmt.Sprintf("%d", d.CurrentCount),
				d.Status,
			})
		}
		table.Render()
	},
}

var scaleDeploymentCmd = &cobra.Command{
	Use:   "scale [id] [replicas]",
	Short: "Scale a deployment",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var replicas int
		_, err := fmt.Sscanf(args[1], "%d", &replicas)
		if err != nil {
			fmt.Printf("Error: invalid replica count: %v\n", err)
			return
		}

		client := getClient()
		err = client.ScaleDeployment(args[0], replicas)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("[SUCCESS] Scaling initiated")
	},
}

var deleteDeploymentCmd = &cobra.Command{
	Use:   "rm [id]",
	Short: "Delete a deployment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		err := client.DeleteDeployment(args[0])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("[SUCCESS] Deletion initiated")
	},
}

func init() {
	createDeploymentCmd.Flags().IntP("replicas", "r", 1, "Number of replicas")
	createDeploymentCmd.Flags().StringP("ports", "p", "", "Ports to expose (e.g. 80:80)")

	containerCmd.AddCommand(createDeploymentCmd)
	containerCmd.AddCommand(listDeploymentsCmd)
	containerCmd.AddCommand(scaleDeploymentCmd)
	containerCmd.AddCommand(deleteDeploymentCmd)

}
