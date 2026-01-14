// Package main provides the cloud CLI entrypoint.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const (
	errorFormat = "Error: %v\n"
	detailRow   = "%-15s %v\n"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage managed database instances (RDS)",
}

var dbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all database instances",
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		databases, err := client.ListDatabases()
		if err != nil {
			fmt.Printf(errorFormat, err)
			return
		}

		if outputJSON {
			data, _ := json.MarshalIndent(databases, "", "  ")
			fmt.Println(string(data))
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ID", "NAME", "ENGINE", "VERSION", "STATUS", "PORT"})

		for _, db := range databases {
			id := db.ID
			if len(id) > 8 {
				id = id[:8]
			}

			table.Append([]string{
				id,
				db.Name,
				db.Engine,
				db.Version,
				db.Status,
				fmt.Sprintf("%d", db.Port),
			})
		}
		table.Render()
	},
}

var dbCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new managed database instance",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		engine, _ := cmd.Flags().GetString("engine")
		version, _ := cmd.Flags().GetString("version")
		vpc, _ := cmd.Flags().GetString("vpc")

		var vpcPtr *string
		if vpc != "" {
			vpcPtr = &vpc
		}

		client := getClient()
		db, err := client.CreateDatabase(name, engine, version, vpcPtr)
		if err != nil {
			fmt.Printf(errorFormat, err)
			return
		}

		fmt.Printf("[SUCCESS] Database %s created successfully!\n", name)
		if outputJSON {
			// Mask password for JSON output
			dbCopy := *db
			dbCopy.Password = "********"
			data, _ := json.MarshalIndent(dbCopy, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("ID:       %s\n", db.ID)
			fmt.Printf("Username: %s\n", db.Username)
			fmt.Println("\n SECURITY WARNING: The following password is shown ONCE only.")
			fmt.Println("    Save it securely - it cannot be retrieved later.")
			// nosemgrep: go.lang.security.audit.xss.print-not-escaped.print-not-escaped
			fmt.Printf("Password: %s\n", db.Password)
			fmt.Println("\nFor security reasons, this password will not be displayed again.")
		}
	},
}

var dbShowCmd = &cobra.Command{
	Use:   "show [id/name]",
	Short: "Show detailed database information",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		client := getClient()
		db, err := client.GetDatabase(id)
		if err != nil {
			fmt.Printf(errorFormat, err)
			return
		}

		fmt.Printf("\nDatabase Details\n")
		fmt.Println(strings.Repeat("-", 40))
		fmt.Printf(detailRow, "ID:", db.ID)
		fmt.Printf(detailRow, "Name:", db.Name)
		fmt.Printf(detailRow, "Engine:", db.Engine)
		fmt.Printf(detailRow, "Version:", db.Version)
		fmt.Printf(detailRow, "Status:", db.Status)
		fmt.Printf(detailRow, "Port:", db.Port)
		fmt.Printf(detailRow, "Username:", db.Username)
		fmt.Printf(detailRow, "VPC ID:", db.VpcID)
		fmt.Printf(detailRow, "Created At:", db.CreatedAt)
		fmt.Println(strings.Repeat("-", 40))
		fmt.Println("")
	},
}

var dbRmCmd = &cobra.Command{
	Use:   "rm [id]",
	Short: "Remove a database instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		client := getClient()
		if err := client.DeleteDatabase(id); err != nil {
			fmt.Printf(errorFormat, err)
			return
		}
		fmt.Println("[SUCCESS] Database removed.")
	},
}

var dbConnCmd = &cobra.Command{
	Use:   "connection [id]",
	Short: "Get database connection string",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		client := getClient()
		connStr, err := client.GetDatabaseConnectionString(id)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Connection String: %s\n", connStr)
	},
}

func init() {
	dbCmd.AddCommand(dbListCmd)
	dbCmd.AddCommand(dbCreateCmd)
	dbCmd.AddCommand(dbShowCmd)
	dbCmd.AddCommand(dbRmCmd)
	dbCmd.AddCommand(dbConnCmd)

	dbCreateCmd.Flags().StringP("name", "n", "", "Name of the database (required)")
	dbCreateCmd.Flags().StringP("engine", "e", "postgres", "Database engine (postgres/mysql)")
	dbCreateCmd.Flags().StringP("version", "v", "16", "Engine version")
	dbCreateCmd.Flags().StringP("vpc", "V", "", "VPC ID to attach to")
	dbCreateCmd.MarkFlagRequired("name")
}
