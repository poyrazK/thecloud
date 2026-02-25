// Package main provides the cloud CLI entrypoint.
package main

import (
	"fmt"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var instanceTypeCmd = &cobra.Command{
	Use:   "instance-type",
	Short: "Manage instance types and pricing",
}

var instanceTypeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available instance types",
	Run: func(cmd *cobra.Command, args []string) {
		client := createClient()
		types, err := client.ListInstanceTypes()
		if err != nil {
			fmt.Printf("Error: %v
", err)
			return
		}

		if jsonOutput {
			printJSON(types)
			return
		}

		if len(types) == 0 {
			fmt.Println("No instance types found.")
			return
		}

		table := tablewriter.NewWriter(cmd.OutOrStdout())
		table.Header([]string{"ID", "NAME", "VCPUS", "MEMORY", "DISK", "PRICE/HR", "CATEGORY"})

		for _, t := range types {
			table.Append([]string{
				t.ID,
				t.Name,
				fmt.Sprintf("%d", t.VCPUs),
				fmt.Sprintf("%d MB", t.MemoryMB),
				fmt.Sprintf("%d GB", t.DiskGB),
				fmt.Sprintf("$%.4f", t.PricePerHr),
				t.Category,
			})
		}
		table.Render()
	},
}

func init() {
	instanceTypeCmd.AddCommand(instanceTypeListCmd)
}
