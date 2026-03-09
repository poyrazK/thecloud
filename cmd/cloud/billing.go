// Package main provides the cloud CLI entrypoint.
package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var billingCmd = &cobra.Command{
	Use:   "billing",
	Short: "View billing and usage information",
}

var billingSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Get billing summary for the current period",
	Run: func(cmd *cobra.Command, args []string) {
		client := createClient(opts)
		summary, err := client.GetBillingSummary(cmd.Context(), nil, nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if opts.JSON {
			printJSON(summary)
			return
		}

		fmt.Printf("Billing Summary (%s to %s)\n", summary.PeriodStart.Format("2006-01-02"), summary.PeriodEnd.Format("2006-01-02"))
		fmt.Printf("Total Amount: %s %.2f\n", summary.Currency, summary.TotalAmount)
		fmt.Println("\nUsage by Type:")
		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"RESOURCE TYPE", "AMOUNT"})
		for t, amt := range summary.UsageByType {
			table.Append([]string{string(t), fmt.Sprintf("%.2f", amt)})
		}
		table.Render()
	},
}

var billingUsageCmd = &cobra.Command{
	Use:   "usage",
	Short: "List detailed usage records",
	Run: func(cmd *cobra.Command, args []string) {
		client := createClient(opts)
		records, err := client.ListUsageRecords(cmd.Context(), nil, nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if opts.JSON {
			printJSON(records)
			return
		}

		if len(records) == 0 {
			fmt.Println("No usage records found.")
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"RESOURCE ID", "TYPE", "QUANTITY", "UNIT", "START TIME"})

		for _, r := range records {
			table.Append([]string{
				truncateID(r.ResourceID.String()),
				string(r.ResourceType),
				fmt.Sprintf("%.2f", r.Quantity),
				r.Unit,
				r.StartTime.Format("2006-01-02 15:04"),
			})
		}
		table.Render()
	},
}

func init() {
	billingCmd.AddCommand(billingSummaryCmd)
	billingCmd.AddCommand(billingUsageCmd)
}
