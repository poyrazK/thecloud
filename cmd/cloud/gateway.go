// Package main provides the cloud CLI entrypoint.
package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

const gatewayErrorFormat = "Error: %v\n"

var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "Manage CloudGateway (API Routes)",
}

var createRouteCmd = &cobra.Command{
	Use:   "create-route [name] [pattern] [target]",
	Short: "Create a new gateway route",
	Long: `Create a gateway route with pattern matching support.

Pattern Syntax:
  - /api/v1/*              Wildcard matching
  - /users/{id}            Named parameter
  - /users/{id:[0-9]+}     Parameter with regex
  - /files/*.{ext}         Named wildcard

Examples:
  cloud gateway create-route users-api "/users/{id}" http://user-service:8080
  cloud gateway create-route files "/files/*" http://storage:8080 --strip`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		strip, _ := cmd.Flags().GetBool("strip")
		limit, _ := cmd.Flags().GetInt("rate-limit")

		client := getClient()
		route, err := client.CreateGatewayRoute(args[0], args[1], args[2], strip, limit)
		if err != nil {
			fmt.Printf(gatewayErrorFormat, err)
			return
		}

		fmt.Printf("[SUCCESS] Route created: %s (Prefix: %s -> %s)\n", route.Name, route.PathPrefix, route.TargetURL)
	},
}

var listRoutesCmd = &cobra.Command{
	Use:   "list-routes",
	Short: "List all gateway routes",
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		routes, err := client.ListGatewayRoutes()
		if err != nil {
			fmt.Printf(gatewayErrorFormat, err)
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ID", "NAME", "PATTERN", "TARGET", "STRIP"})
		for _, r := range routes {
			_ = table.Append([]string{r.ID, r.Name, r.PathPattern, r.TargetURL, fmt.Sprintf("%v", r.StripPrefix)})
		}
		_ = table.Render()
	},
}

var deleteRouteCmd = &cobra.Command{
	Use:   "rm-route [id]",
	Short: "Delete a gateway route",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		err := client.DeleteGatewayRoute(args[0])
		if err != nil {
			fmt.Printf(gatewayErrorFormat, err)
			return
		}
		fmt.Println("[SUCCESS] Route deleted")
	},
}

func init() {
	createRouteCmd.Flags().Bool("strip", true, "Strip prefix from target request")
	createRouteCmd.Flags().Int("rate-limit", 100, "Rate limit (req/sec)")

	gatewayCmd.AddCommand(createRouteCmd)
	gatewayCmd.AddCommand(listRoutesCmd)
	gatewayCmd.AddCommand(deleteRouteCmd)

}
