package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var subnetCmd = &cobra.Command{
	Use:   "subnet",
	Short: "Manage Subnets within VPCs",
}

var subnetCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new subnet in a VPC",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID, _ := cmd.Flags().GetString("vpc-id")
		cidr, _ := cmd.Flags().GetString("cidr-block")
		az, _ := cmd.Flags().GetString("az")

		if vpcID == "" {
			fmt.Println("Error: --vpc-id is required")
			return
		}

		req := map[string]interface{}{
			"name":              args[0],
			"cidr_block":        cidr,
			"availability_zone": az,
		}

		data, err := json.Marshal(req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		resp, err := makeRequest(http.MethodPost, fmt.Sprintf("/vpcs/%s/subnets", vpcID), data)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer resp.Body.Close()

		handleResponse(resp, "Subnet created")
	},
}

var subnetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List subnets in a VPC",
	Run: func(cmd *cobra.Command, args []string) {
		vpcID, _ := cmd.Flags().GetString("vpc-id")
		if vpcID == "" {
			fmt.Println("Error: --vpc-id is required")
			return
		}

		resp, err := makeRequest(http.MethodGet, fmt.Sprintf("/vpcs/%s/subnets", vpcID), nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			handleResponse(resp, "")
			return
		}

		var result struct {
			Data []map[string]interface{} `json:"data"`
		}
		json.NewDecoder(resp.Body).Decode(&result)

		if outputJSON {
			data, _ := json.MarshalIndent(result.Data, "", "  ")
			fmt.Println(string(data))
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ID", "NAME", "CIDR", "AZ", "GATEWAY"})

		for _, s := range result.Data {
			table.Append([]string{
				s["id"].(string)[:8],
				s["name"].(string),
				s["cidr_block"].(string),
				s["availability_zone"].(string),
				s["gateway_ip"].(string),
			})
		}
		table.Render()
	},
}

var subnetRmCmd = &cobra.Command{
	Use:   "rm [id]",
	Short: "Remove a subnet",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := makeRequest(http.MethodDelete, fmt.Sprintf("/subnets/%s", args[0]), nil)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer resp.Body.Close()

		handleResponse(resp, "Subnet removed")
	},
}

func init() {
	subnetCreateCmd.Flags().String("vpc-id", "", "VPC ID")
	subnetCreateCmd.Flags().String("cidr-block", "", "CIDR block (e.g. 10.0.1.0/24)")
	subnetCreateCmd.Flags().String("az", "us-east-1a", "Availability Zone")
	subnetCreateCmd.MarkFlagRequired("vpc-id")
	subnetCreateCmd.MarkFlagRequired("cidr-block")

	subnetListCmd.Flags().String("vpc-id", "", "VPC ID")
	subnetListCmd.MarkFlagRequired("vpc-id")

	subnetCmd.AddCommand(subnetCreateCmd, subnetListCmd, subnetRmCmd)
	rootCmd.AddCommand(subnetCmd)
}
