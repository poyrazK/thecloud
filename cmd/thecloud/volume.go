package main

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Manage block storage volumes",
}

var volumeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all volumes",
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		volumes, err := client.ListVolumes()
		if err != nil {
			printError(err)
		}

		if outputJSON {
			printJSON(volumes)
			return
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ID", "NAME", "SIZE", "STATUS", "ATTACHED TO"})

		for _, v := range volumes {
			id := v.ID.String()
			if len(id) > 8 {
				id = id[:8]
			}
			attachedTo := "-"
			if v.InstanceID != nil {
				attachedTo = v.InstanceID.String()[:8]
			}
			table.Append([]string{
				id,
				v.Name,
				fmt.Sprintf("%d GB", v.SizeGB),
				v.Status,
				attachedTo,
			})
		}
		table.Render()
	},
}

var volumeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new volume",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		size, _ := cmd.Flags().GetInt("size")

		client := getClient()
		vol, err := client.CreateVolume(name, size)
		if err != nil {
			printError(err)
		}

		printDataOrStatus(vol, "[SUCCESS] Volume created!")
	},
}

var volumeDeleteCmd = &cobra.Command{
	Use:   "rm [id/name]",
	Short: "Delete a volume",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		client := getClient()
		if err := client.DeleteVolume(id); err != nil {
			printError(err)
		}
		printStatus(fmt.Sprintf("[SUCCESS] Volume %s deleted.", id))
	},
}

func init() {
	rootCmd.AddCommand(volumeCmd)
	volumeCmd.AddCommand(volumeListCmd)
	volumeCmd.AddCommand(volumeCreateCmd)
	volumeCmd.AddCommand(volumeDeleteCmd)

	volumeCreateCmd.Flags().StringP("name", "n", "", "Name of the volume (required)")
	volumeCreateCmd.Flags().IntP("size", "s", 1, "Size in GB")
	volumeCreateCmd.MarkFlagRequired("name")
}
