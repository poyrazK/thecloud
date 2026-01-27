// Package main provides the cloud CLI entrypoint.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const loadBalancerErrorFormat = "Error: %v\n"

var lbCmd = &cobra.Command{
	Use:   "lb",
	Short: "Manage Load Balancers",
}

var lbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all load balancers",
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		lbs, err := client.ListLBs()
		if err != nil {
			fmt.Printf(loadBalancerErrorFormat, err)
			return
		}

		if outputJSON {
			data, _ := json.MarshalIndent(lbs, "", "  ")
			fmt.Println(string(data))
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ID", "NAME", "VPC ID", "PORT", "ALGO", "STATUS"})

		for _, v := range lbs {
			table.Append([]string{
				v.ID[:8],
				v.Name,
				v.VpcID[:8],
				fmt.Sprintf("%d", v.Port),
				v.Algorithm,
				string(v.Status),
			})
		}
		table.Render()
	},
}

var lbCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new load balancer",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		vpcID, _ := cmd.Flags().GetString("vpc")
		port, _ := cmd.Flags().GetInt("port")
		algo, _ := cmd.Flags().GetString("algorithm")

		client := getClient()
		lb, err := client.CreateLB(name, vpcID, port, algo)
		if err != nil {
			fmt.Printf(loadBalancerErrorFormat, err)
			return
		}

		fmt.Printf("[SUCCESS] Load Balancer %s creation initiated!\n", lb.Name)
		fmt.Printf("ID: %s\n", lb.ID)
		fmt.Printf("Status: %s (It will be ACTIVE shortly)\n", lb.Status)
	},
}

var lbRmCmd = &cobra.Command{
	Use:   "rm [id]",
	Short: "Remove a load balancer",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		client := getClient()
		if err := client.DeleteLB(id); err != nil {
			fmt.Printf(loadBalancerErrorFormat, err)
			return
		}

		fmt.Printf("[SUCCESS] Load Balancer %s deletion initiated.\n", id)
	},
}

var lbAddTargetCmd = &cobra.Command{
	Use:   "add-target [lb-id]",
	Short: "Add a target instance to a load balancer",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		lbID := args[0]
		instID, _ := cmd.Flags().GetString("instance")
		port, _ := cmd.Flags().GetInt("port")
		weight, _ := cmd.Flags().GetInt("weight")

		client := getClient()
		if err := client.AddLBTarget(lbID, instID, port, weight); err != nil {
			fmt.Printf(loadBalancerErrorFormat, err)
			return
		}

		fmt.Printf("[SUCCESS] Target %s added to LB %s.\n", instID, lbID)
	},
}

var lbRemoveTargetCmd = &cobra.Command{
	Use:   "rm-target [lb-id] [instance-id]",
	Short: "Remove a target instance from a load balancer",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		lbID := args[0]
		instID := args[1]
		client := getClient()
		if err := client.RemoveLBTarget(lbID, instID); err != nil {
			fmt.Printf(loadBalancerErrorFormat, err)
			return
		}

		fmt.Printf("[SUCCESS] Target %s removed from LB %s.\n", instID, lbID)
	},
}

func init() {
	lbCreateCmd.Flags().String("name", "", "Name of the load balancer")
	lbCreateCmd.MarkFlagRequired("name")
	lbCreateCmd.Flags().String("vpc", "", "VPC ID")
	lbCreateCmd.MarkFlagRequired("vpc")
	lbCreateCmd.Flags().Int("port", 80, "Public port for the LB")
	lbCreateCmd.Flags().String("algorithm", "round-robin", "LB algorithm (round-robin or least-conn)")

	lbAddTargetCmd.Flags().String("instance", "", "Target instance ID")
	lbAddTargetCmd.MarkFlagRequired("instance")
	lbAddTargetCmd.Flags().Int("port", 80, "Port on the instance")
	lbAddTargetCmd.MarkFlagRequired("port")
	lbAddTargetCmd.Flags().Int("weight", 1, "Weight for the target (optional)")

	lbCmd.AddCommand(lbListCmd)
	lbCmd.AddCommand(lbCreateCmd)
	lbCmd.AddCommand(lbRmCmd)
	lbCmd.AddCommand(lbAddTargetCmd)
	lbCmd.AddCommand(lbRemoveTargetCmd)
	lbCmd.AddCommand(lbListTargetsCmd)
}

var lbListTargetsCmd = &cobra.Command{
	Use:   "targets <lb-id>",
	Short: "List targets for a load balancer",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		targets, err := client.ListLBTargets(args[0])
		if err != nil {
			fmt.Printf(loadBalancerErrorFormat, err)
			return
		}

		if outputJSON {
			data, _ := json.MarshalIndent(targets, "", "  ")
			fmt.Println(string(data))
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"INSTANCE ID", "PORT", "WEIGHT", "HEALTH"})
		for _, t := range targets {
			id := t.InstanceID
			if len(id) > 8 {
				id = id[:8]
			}
			table.Append([]string{
				id,
				fmt.Sprintf("%d", t.Port),
				fmt.Sprintf("%d", t.Weight),
				t.Health,
			})
		}
		table.Render()
	},
}
