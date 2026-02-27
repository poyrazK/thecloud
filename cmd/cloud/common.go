package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/poyrazk/thecloud/pkg/sdk"
)

type CLIOptions struct {
	JSON     bool
	APIKey   string
	APIURL   string
	TenantID string
}

var opts CLIOptions

func createClient(o CLIOptions) *sdk.Client {
	key := o.APIKey
	if key == "" {
		key = os.Getenv("CLOUD_API_KEY")
	}
	if key == "" {
		key = loadConfig()
	}

	if key == "" {
		fmt.Println("[WARN] No API Key found. Run 'cloud auth create-demo <name>' to get one.")
		os.Exit(1)
	}

	client := sdk.NewClient(o.APIURL, key)

	tenant := o.TenantID
	if tenant == "" {
		tenant = os.Getenv("CLOUD_TENANT_ID")
	}

	if tenant != "" {
		client.SetTenant(tenant)
	}
	return client
}

func printJSON(data interface{}) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(b))
}

func truncateID(id string) string {
	const n = 8
	if len(id) <= n {
		return id
	}
	return id[:n]
}
