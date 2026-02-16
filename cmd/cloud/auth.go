// Package main provides the cloud CLI entrypoint.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/poyrazk/thecloud/pkg/sdk"
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
		client := sdk.NewClient(apiURL, "") // Key not needed for creation usually
		key, err := client.CreateKey(name)

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("[INFO] Generated Key: %s\n", key)
		saveConfig(key)
		fmt.Println("[SUCCESS] Key saved to configuration. You can now use 'cloud' commands without flags.")
	},
}

var loginCmd = &cobra.Command{
	Use:   "login [key]",
	Short: "Save an existing API key to configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		saveConfig(key)
		fmt.Println("[SUCCESS] Key saved to configuration.")
	},
}

// Config persistence
type Config struct {
	APIKey string `json:"api_key"`
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cloud", "config.json")
}

func saveConfig(key string) {
	path := getConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		fmt.Printf("Warning: failed to create config directory: %v\n", err)
	}

	cfg := Config{APIKey: key}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		fmt.Printf("Error: failed to marshal config: %v\n", err)
		return
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		fmt.Printf("Error: failed to write config file: %v\n", err)
	}
}

func loadConfig() string {
	path := getConfigPath()
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return ""
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return ""
	}
	return cfg.APIKey
}

func init() {
	authCmd.AddCommand(createDemoCmd)
	authCmd.AddCommand(loginCmd)
}
