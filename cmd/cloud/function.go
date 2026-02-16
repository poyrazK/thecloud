// Package main provides the cloud CLI entrypoint.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var functionCmd = &cobra.Command{
	Use:   "function",
	Short: "Manage CloudFunctions",
}

var createFnCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new function",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		runtime, _ := cmd.Flags().GetString("runtime")
		handler, _ := cmd.Flags().GetString("handler")
		codePath, _ := cmd.Flags().GetString("code")

		code, err := os.ReadFile(filepath.Clean(codePath))
		if err != nil {
			return fmt.Errorf("failed to read code file: %w", err)
		}

		client := getClient()
		fn, err := client.CreateFunction(name, runtime, handler, code)
		if err != nil {
			return err
		}

		fmt.Printf("Function %s created successfully (ID: %s)\n", fn.Name, fn.ID)
		return nil
	},
}

var listFnCmd = &cobra.Command{
	Use:   "list",
	Short: "List all functions",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		functions, err := client.ListFunctions()
		if err != nil {
			return err
		}

		if len(functions) == 0 {
			fmt.Println("No functions found.")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ID", "Name", "Runtime", "Status", "Created At"})
		for _, f := range functions {
			_ = table.Append([]string{f.ID, f.Name, f.Runtime, f.Status, f.CreatedAt.Format("2006-01-02 15:04:05")})
		}
		_ = table.Render()
		return nil
	},
}

var invokeFnCmd = &cobra.Command{
	Use:   "invoke [name/id]",
	Short: "Invoke a function",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		payloadStr, _ := cmd.Flags().GetString("payload")
		payloadFile, _ := cmd.Flags().GetString("payload-file")
		async, _ := cmd.Flags().GetBool("async")

		var payload []byte
		var err error
		if payloadFile != "" {
			payload, err = os.ReadFile(filepath.Clean(payloadFile))
			if err != nil {
				return err
			}
		} else {
			payload = []byte(payloadStr)
		}

		client := getClient()

		// Map name to ID if needed
		targetID := id
		functions, err := client.ListFunctions()
		if err == nil {
			for _, f := range functions {
				if f.Name == id {
					targetID = f.ID
					break
				}
			}
		}

		invocation, err := client.InvokeFunction(targetID, payload, async)
		if err != nil {
			return err
		}

		if async {
			fmt.Printf("Invocation started (ID: %s)\n", invocation.ID)
		} else {
			fmt.Printf("Status: %s\n", invocation.Status)
			fmt.Printf("Exit Code: %d\n", invocation.StatusCode)
			fmt.Printf("Duration: %dms\n", invocation.DurationMs)
			fmt.Println("\nLogs:")
			fmt.Println(invocation.Logs)
		}
		return nil
	},
}

var logsFnCmd = &cobra.Command{
	Use:   "logs [name/id]",
	Short: "Get recent logs for a function",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		client := getClient()

		targetID := id
		functions, err := client.ListFunctions()
		if err == nil {
			for _, f := range functions {
				if f.Name == id {
					targetID = f.ID
					break
				}
			}
		}

		invocations, err := client.GetFunctionLogs(targetID)
		if err != nil {
			return err
		}

		for _, i := range invocations {
			fmt.Printf("--- Invocation %s (%s) ---\n", i.ID, i.StartedAt.Format("15:04:05"))
			fmt.Printf("Status: %s, Exit Code: %d, Duration: %dms\n", i.Status, i.StatusCode, i.DurationMs)
			fmt.Println(i.Logs)
			fmt.Println()
		}
		return nil
	},
}

var rmFnCmd = &cobra.Command{
	Use:   "rm [name/id]",
	Short: "Remove a function",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		client := getClient()

		targetID := id
		functions, err := client.ListFunctions()
		if err == nil {
			for _, f := range functions {
				if f.Name == id {
					targetID = f.ID
					break
				}
			}
		}

		if err := client.DeleteFunction(targetID); err != nil {
			return err
		}
		fmt.Printf("Function %s removed.\n", id)
		return nil
	},
}

func init() {
	createFnCmd.Flags().StringP("name", "n", "", "Function name")
	createFnCmd.Flags().StringP("runtime", "r", "nodejs20", "Runtime (nodejs20, python312, go122, ruby33, java21)")
	createFnCmd.Flags().StringP("handler", "H", "index.handler", "Handler name")
	createFnCmd.Flags().StringP("code", "c", "", "Path to code zip file")
	_ = createFnCmd.MarkFlagRequired("name")
	_ = createFnCmd.MarkFlagRequired("code")

	invokeFnCmd.Flags().StringP("payload", "p", "{}", "JSON payload")
	invokeFnCmd.Flags().StringP("payload-file", "f", "", "Path to payload file")
	invokeFnCmd.Flags().BoolP("async", "a", false, "Invoke asynchronously")

	functionCmd.AddCommand(createFnCmd)
	functionCmd.AddCommand(listFnCmd)
	functionCmd.AddCommand(invokeFnCmd)
	functionCmd.AddCommand(logsFnCmd)
	functionCmd.AddCommand(rmFnCmd)
}
