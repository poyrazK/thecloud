// Package main provides the cloud CLI entrypoint.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "View platform audit logs",
}

var auditListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent audit logs",
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")

		client := createClient()
		logs, err := client.ListAuditLogs(limit)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if jsonOutput {
			printJSON(logs)
			return
		}

		if len(logs) == 0 {
			fmt.Println("No audit logs found.")
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"TIMESTAMP", "ACTION", "RESOURCE", "IP ADDRESS"})

		for _, l := range logs {
			resource := l.ResourceType
			if l.ResourceID != "" {
				shortID := l.ResourceID
				if len(shortID) > 8 {
					shortID = shortID[:8]
				}
				resource = fmt.Sprintf("%s/%s", l.ResourceType, shortID)
			}

			_ = table.Append([]string{
				l.CreatedAt.Format(time.RFC3339),
				l.Action,
				resource,
				l.IPAddress,
			})
		}
		_ = table.Render()
	},
}

func init() {
	auditListCmd.Flags().Int("limit", 50, "Limit number of logs")
	auditCmd.AddCommand(auditListCmd)
}
