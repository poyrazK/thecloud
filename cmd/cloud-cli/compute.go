package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var apiURL = "http://localhost:8080"
var outputJSON bool

var computeCmd = &cobra.Command{
	Use:   "compute",
	Short: "Manage compute instances",
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all instances",
	Run: func(cmd *cobra.Command, args []string) {
		client := resty.New()
		resp, err := client.R().Get(apiURL + "/instances")
		if err != nil {
			fmt.Printf("Error connecting to API: %v\n", err)
			return
		}

		if outputJSON {
			fmt.Println(string(resp.Body()))
			return
		}

		var result struct {
			Data []map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body(), &result); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ID", "NAME", "IMAGE", "STATUS", "CREATED AT"})

		for _, inst := range result.Data {
			id := fmt.Sprintf("%v", inst["id"])
			if len(id) > 8 {
				id = id[:8]
			}
			table.Append([]string{
				id,
				fmt.Sprintf("%v", inst["name"]),
				fmt.Sprintf("%v", inst["image"]),
				fmt.Sprintf("%v", inst["status"]),
				fmt.Sprintf("%v", inst["created_at"]),
			})
		}
		table.Render()
	},
}

var launchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Launch a new instance",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		image, _ := cmd.Flags().GetString("image")

		client := resty.New()
		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(map[string]string{
				"name":  name,
				"image": image,
			}).
			Post(apiURL + "/instances")

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if resp.IsError() {
			fmt.Printf("Failed: %s\n", resp.String())
			return
		}

		fmt.Printf("🚀 Instance launched successfully!\n%s\n", resp.String())
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop [id]",
	Short: "Stop an instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		client := resty.New()
		resp, err := client.R().Post(apiURL + "/instances/" + id + "/stop")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if resp.IsError() {
			fmt.Printf("Failed: %s\n", resp.String())
			return
		}

		fmt.Println("🛑 Instance stop initiated.")
	},
}

func init() {
	computeCmd.AddCommand(listCmd)
	computeCmd.AddCommand(launchCmd)
	computeCmd.AddCommand(stopCmd)

	launchCmd.Flags().StringP("name", "n", "", "Name of the instance (required)")
	launchCmd.Flags().StringP("image", "i", "alpine", "Image to use")
	launchCmd.MarkFlagRequired("name")

	rootCmd.PersistentFlags().BoolVarP(&outputJSON, "json", "j", false, "Output in JSON format")
	rootCmd.AddCommand(computeCmd)
}
