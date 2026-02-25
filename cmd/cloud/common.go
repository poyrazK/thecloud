package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/poyrazk/thecloud/pkg/sdk"
)

var (
	apiURL     = "http://localhost:8080"
	jsonOutput bool
	apiKey     string
	tenantID   string
)

func createClient() *sdk.Client {
	key := apiKey
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

	client := sdk.NewClient(apiURL, key)

	tenant := tenantID
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
