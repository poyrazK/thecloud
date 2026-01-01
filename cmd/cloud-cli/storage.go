package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Manage object storage",
}

var storageListCmd = &cobra.Command{
	Use:   "list [bucket]",
	Short: "List objects in a bucket",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		bucket := args[0]
		client := getClient()
		resp, err := client.R().Get(apiURL + "/storage/" + bucket)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
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
		table.Header([]string{"KEY", "SIZE", "CREATED AT", "ARN"})

		for _, obj := range result.Data {
			table.Append([]string{
				fmt.Sprintf("%v", obj["key"]),
				fmt.Sprintf("%v", obj["size_bytes"]),
				fmt.Sprintf("%v", obj["created_at"]),
				fmt.Sprintf("%v", obj["arn"]),
			})
		}
		table.Render()
	},
}

var storageUploadCmd = &cobra.Command{
	Use:   "upload [bucket] [file]",
	Short: "Upload a file to a bucket",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		bucket := args[0]
		filePath := args[1]
		key := cmd.Flag("key").Value.String()
		if key == "" {
			key = filePath // Default key to filename
		}

		f, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return
		}
		defer f.Close()

		client := getClient()
		resp, err := client.R().
			SetBody(f).
			Put(apiURL + "/storage/" + bucket + "/" + key)

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if resp.IsError() {
			fmt.Printf("Failed: %s\n", resp.String())
			return
		}

		fmt.Printf("✅ Uploaded %s to bucket %s\n", key, bucket)
	},
}

var storageDownloadCmd = &cobra.Command{
	Use:   "download [bucket] [key] [dest]",
	Short: "Download an object from a bucket",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		bucket := args[0]
		key := args[1]
		dest := args[2]

		client := getClient()
		resp, err := client.R().
			SetDoNotParseResponse(true).
			Get(apiURL + "/storage/" + bucket + "/" + key)

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer resp.RawBody().Close()

		if resp.IsError() {
			fmt.Printf("Failed: Status %d\n", resp.StatusCode())
			return
		}

		out, err := os.Create(dest)
		if err != nil {
			fmt.Printf("Error creating destination file: %v\n", err)
			return
		}
		defer out.Close()

		_, err = io.Copy(out, resp.RawBody())
		if err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			return
		}

		fmt.Printf("✅ Downloaded %s to %s\n", key, dest)
	},
}

var storageDeleteCmd = &cobra.Command{
	Use:   "delete [bucket] [key]",
	Short: "Delete an object from a bucket",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		bucket := args[0]
		key := args[1]

		client := getClient()
		resp, err := client.R().Delete(apiURL + "/storage/" + bucket + "/" + key)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if resp.IsError() {
			fmt.Printf("Failed: %s\n", resp.String())
			return
		}

		fmt.Printf("🗑️ Deleted %s from bucket %s\n", key, bucket)
	},
}

func init() {
	storageCmd.AddCommand(storageListCmd)
	storageCmd.AddCommand(storageUploadCmd)
	storageCmd.AddCommand(storageDownloadCmd)
	storageCmd.AddCommand(storageDeleteCmd)

	storageUploadCmd.Flags().String("key", "", "Custom key for the object")
}
