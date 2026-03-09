// Package main provides the cloud CLI entrypoint.
package main

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/spf13/cobra"
)

const rbacErrorFormat = "Error: %v\n"

var rolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "Manage RBAC roles and bindings",
}

var createRoleCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new role",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		desc, err := cmd.Flags().GetString("description")
		if err != nil {
			fmt.Printf("Error getting description flag: %v\n", err)
			return
		}
		permsStr, err := cmd.Flags().GetString("permissions")
		if err != nil {
			fmt.Printf("Error getting permissions flag: %v\n", err)
			return
		}

		var permissions []domain.Permission
		if permsStr != "" {
			for _, p := range strings.Split(permsStr, ",") {
				permissions = append(permissions, domain.Permission(strings.TrimSpace(p)))
			}
		}

		client := createClient(opts)
		role, err := client.CreateRole(name, desc, permissions)
		if err != nil {
			fmt.Printf(rbacErrorFormat, err)
			return
		}

		fmt.Printf("[SUCCESS] Role created: %s (%s)\n", role.Name, role.ID)
	},
}

var listRolesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all roles",
	Run: func(cmd *cobra.Command, args []string) {
		client := createClient(opts)
		roles, err := client.ListRoles()
		if err != nil {
			fmt.Printf(rbacErrorFormat, err)
			return
		}

		fmt.Printf("%-36s %-20s %s\n", "ID", "NAME", "PERMISSIONS")
		for _, r := range roles {
			perms := []string{}
			for _, p := range r.Permissions {
				perms = append(perms, string(p))
			}
			fmt.Printf("%-36s %-20s %s\n", r.ID, r.Name, strings.Join(perms, ", "))
		}
	},
}

var bindRoleCmd = &cobra.Command{
	Use:   "bind [user-id-or-email] [role-name]",
	Short: "Assign a role to a user",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		userIdentifier := args[0]
		roleName := args[1]

		client := createClient(opts)
		if err := client.BindRole(userIdentifier, roleName); err != nil {
			fmt.Printf(rbacErrorFormat, err)
			return
		}

		fmt.Printf("[SUCCESS] Role %s bound to user %s\n", roleName, userIdentifier)
	},
}

var listBindingsCmd = &cobra.Command{
	Use:   "list-bindings",
	Short: "List all role bindings",
	Run: func(cmd *cobra.Command, args []string) {
		client := createClient(opts)
		users, err := client.ListRoleBindings()
		if err != nil {
			fmt.Printf(rbacErrorFormat, err)
			return
		}

		fmt.Printf("%-36s %-30s %s\n", "USER ID", "EMAIL", "ROLE")
		for _, u := range users {
			fmt.Printf("%-36s %-30s %s\n", u.ID, u.Email, u.Role)
		}
	},
}

var deleteRoleCmd = &cobra.Command{
	Use:     "rm [id]",
	Aliases: []string{"delete"},
	Short:   "Delete a role",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := uuid.Parse(args[0])
		if err != nil {
			fmt.Printf("Error: invalid role ID: %v\n", err)
			return
		}

		client := createClient(opts)
		if err := client.DeleteRole(id); err != nil {
			fmt.Printf(rbacErrorFormat, err)
			return
		}

		fmt.Printf("[SUCCESS] Role %s deleted\n", id)
	},
}

func init() {
	createRoleCmd.Flags().String("description", "", "Role description")
	createRoleCmd.Flags().String("permissions", "", "Comma-separated list of permissions")

	rolesCmd.AddCommand(createRoleCmd)
	rolesCmd.AddCommand(listRolesCmd)
	rolesCmd.AddCommand(bindRoleCmd)
	rolesCmd.AddCommand(listBindingsCmd)
	rolesCmd.AddCommand(deleteRoleCmd)
}
