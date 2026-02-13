package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var sshKeyCmd = &cobra.Command{
	Use:   "ssh-key",
	Short: "Manage SSH public keys",
}

var registerKeyCmd = &cobra.Command{
	Use:   "register [name] [public_key_file]",
	Short: "Register a new SSH public key",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		keyFile := args[1]

		keyData, err := os.ReadFile(keyFile)
		if err != nil {
			fmt.Printf("Error reading key file: %v\n", err)
			return
		}

		client := getClient()
		key, err := client.RegisterSSHKey(name, string(keyData))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("[SUCCESS] SSH Key registered.")
		fmt.Printf("ID: %s\n", key.ID)
	},
}

var listKeysCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered SSH keys",
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		keys, err := client.ListSSHKeys()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		for _, k := range keys {
			fmt.Printf("%-36s %s\n", k.ID, k.Name)
		}
	},
}

func init() {
	sshKeyCmd.AddCommand(registerKeyCmd)
	sshKeyCmd.AddCommand(listKeysCmd)
	rootCmd.AddCommand(sshKeyCmd)
}
