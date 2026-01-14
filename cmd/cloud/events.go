// Package main provides the cloud CLI entrypoint.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Manage system events (audit logs)",
}

var listEventsCmd = &cobra.Command{
	Use:   "list",
	Short: "List system events",
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		events, err := client.ListEvents()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if outputJSON {
			data, _ := json.MarshalIndent(events, "", "  ")
			fmt.Println(string(data))
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"TIME", "ACTION", "RESOURCE", "TYPE", "DETAILS"})

		for _, e := range events {
			// Format time
			val := e.CreatedAt.Format(time.RFC3339)

			// Simple metadata formatting
			meta := string(e.Metadata)
			if len(meta) > 50 {
				meta = meta[:47] + "..."
			}

			table.Append([]string{
				val,
				e.Action,
				e.ResourceID,
				e.ResourceType,
				meta,
			})
		}
		table.Render()
	},
}

func init() {
	eventsCmd.AddCommand(listEventsCmd)
}
