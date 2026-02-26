// Package main provides the cloud CLI entrypoint.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const (
	defaultAuditLimit = 50
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "View platform audit logs",
}

var auditListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent audit logs",
	Run: func(cmd *cobra.Command, args []string) {
		limit, err := cmd.Flags().GetInt("limit")
		if err != nil {
			fmt.Printf("Error parsing limit: %v\n", err)
			return
		}
		if limit <= 0 {
			fmt.Println("Error: limit must be a positive integer")
			return
		}

		client := createClient(opts)
		logs, err := client.ListAuditLogs(cmd.Context(), limit)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if opts.JSON {
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

			if err := table.Append([]string{
				l.CreatedAt.Format(time.RFC3339),
				l.Action,
				resource,
				l.IPAddress,
			}); err != nil {
				fmt.Printf("Error appending to table: %v\n", err)
				return
			}
		}
		if err := table.Render(); err != nil {
			fmt.Printf("Error rendering table: %v\n", err)
		}
	},
}

func init() {
	auditListCmd.Flags().Int("limit", defaultAuditLimit, "Limit number of logs")
	auditCmd.AddCommand(auditListCmd)
}
