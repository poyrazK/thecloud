package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var vpcCmd = &cobra.Command{
	Use:   "vpc",
	Short: "Manage Virtual Private Clouds (VPCs)",
}

var vpcListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all VPCs",
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		vpcs, err := client.ListVPCs()
		if err != nil {
			printError(err)
		}

		if outputJSON {
			printJSON(vpcs)
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ID", "NAME", "NETWORK ID", "CREATED AT"})

		for _, v := range vpcs {
			table.Append([]string{
				v.ID[:8],
				v.Name,
				v.NetworkID[:12],
				v.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
		table.Render()
	},
}

var vpcCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new VPC",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		client := getClient()
		vpc, err := client.CreateVPC(name)
		if err != nil {
			printError(err)
		}

		printDataOrStatus(vpc, fmt.Sprintf("[SUCCESS] VPC %s created successfully!", vpc.Name))
		if !outputJSON {
			fmt.Printf("ID: %s\n", vpc.ID)
			fmt.Printf("Network ID: %s\n", vpc.NetworkID)
		}
	},
}

var vpcRmCmd = &cobra.Command{
	Use:   "rm [id/name]",
	Short: "Remove a VPC",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		client := getClient()
		if err := client.DeleteVPC(id); err != nil {
			printError(err)
		}

		printStatus(fmt.Sprintf("[SUCCESS] VPC %s removed successfully.", id))
	},
}

func init() {
	vpcCmd.AddCommand(vpcListCmd)
	vpcCmd.AddCommand(vpcCreateCmd)
	vpcCmd.AddCommand(vpcRmCmd)
}
