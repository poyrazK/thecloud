package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var createDemoCmd = &cobra.Command{
	Use:   "create-demo [name]",
	Short: "Generate a demo API key and save it",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		client := resty.New()
		resp, err := client.R().
			SetBody(map[string]string{"name": name}).
			Post(apiURL + "/auth/keys")

		if err != nil {
			fmt.Printf("Error connecting to API: %v\n", err)
			return
		}

		if resp.IsError() {
			fmt.Printf("Failed: %s\n", resp.String())
			return
		}

		var result struct {
			Data struct {
				Key string `json:"key"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body(), &result); err != nil {
			fmt.Printf("Error parsing response: %v\n", err)
			return
		}

		key := result.Data.Key
		fmt.Printf("🔑 Generated Key: %s\n", key)
		saveConfig(key)
		fmt.Println("✅ Key saved to configuration. You can now use 'cloud' commands without flags.")
	},
}

var loginCmd = &cobra.Command{
	Use:   "login [key]",
	Short: "Save an existing API key to configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		saveConfig(key)
		fmt.Println("✅ Key saved to configuration.")
	},
}

// Config persistence
type Config struct {
	APIKey string `json:"api_key"`
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".miniaws", "config.json")
}

func saveConfig(key string) {
	path := getConfigPath()
	os.MkdirAll(filepath.Dir(path), 0755)

	cfg := Config{APIKey: key}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(path, data, 0644)
}

func loadConfig() string {
	path := getConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var cfg Config
	json.Unmarshal(data, &cfg)
	return cfg.APIKey
}

func init() {
	authCmd.AddCommand(createDemoCmd)
	authCmd.AddCommand(loginCmd)
}
