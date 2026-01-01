package main

import (
	"encoding/json"
	"fmt"
	"os"

	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/go-resty/resty/v2"
	"github.com/olekukonko/tablewriter"
)

var apiURL = "http://localhost:8080"
var apiKey string

func main() {
	// 1. Auth Setup
	apiKey = os.Getenv("MINIAWS_API_KEY")
	if apiKey == "" {
		fmt.Println("⚠️  MINIAWS_API_KEY not set.")

		createDemo := false
		prompt := &survey.Confirm{
			Message: "Would you like to generate a temporary key for this session?",
			Default: true,
		}
		survey.AskOne(prompt, &createDemo)

		if createDemo {
			// Auto-generate key
			var name string
			namePrompt := &survey.Input{
				Message: "Enter a name for your key (e.g. demo-user):",
				Default: "demo-user",
			}
			survey.AskOne(namePrompt, &name)

			// Call API to create key
			client := resty.New()
			resp, err := client.R().
				SetBody(map[string]string{"name": name}).
				Post(apiURL + "/auth/keys")

			if err == nil && !resp.IsError() {
				var result struct {
					Data struct {
						Key string `json:"key"`
					} `json:"data"`
				}
				json.Unmarshal(resp.Body(), &result)
				apiKey = result.Data.Key
				fmt.Printf("🔑 Generated Key: %s\n\n", apiKey)
			} else {
				fmt.Println("❌ Failed to generate key. Falling back to manual input.")
			}
		}

		// Fallback: manual input if still empty
		if apiKey == "" {
			manualPrompt := &survey.Input{
				Message: "Enter your API Key:",
			}
			survey.AskOne(manualPrompt, &apiKey)
		}
	}

	for {
		// 2. Main Menu
		mode := ""
		prompt := &survey.Select{
			Message: "☁️  Cloud CLI Control Panel - What would you like to do?",
			Options: []string{"List Instances", "Launch Instance", "Stop Instance", "View Logs", "View Details", "Exit"},
		}
		if err := survey.AskOne(prompt, &mode); err != nil {
			fmt.Println("Bye!")
			return
		}

		// 3. Dispatch
		switch mode {
		case "List Instances":
			listInstances()
		case "Launch Instance":
			launchInstance()
		case "Stop Instance":
			stopInstance()
		case "View Logs":
			viewLogs()
		case "View Details":
			showInstance()
		case "Exit":
			fmt.Println("👋 See you in the cloud!")
			return
		}
		fmt.Println("") // Spacer
	}
}

func getClient() *resty.Client {
	client := resty.New()
	client.SetHeader("X-API-Key", apiKey)
	return client
}

func listInstances() {
	client := getClient()
	resp, err := client.R().Get(apiURL + "/instances")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Clear screen for better UX
	fmt.Print("\033[H\033[2J")

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"ID", "NAME", "IMAGE", "STATUS", "ACCESS"})

	for _, inst := range result.Data {
		id := fmt.Sprintf("%v", inst["id"])

		access := "-"
		ports := fmt.Sprintf("%v", inst["ports"])
		if ports != "" && inst["status"] == "RUNNING" {
			pList := strings.Split(ports, ",")
			var mappings []string
			for _, mapping := range pList {
				parts := strings.Split(mapping, ":")
				if len(parts) == 2 {
					mappings = append(mappings, fmt.Sprintf("localhost:%s->%s", parts[0], parts[1]))
				}
			}
			access = strings.Join(mappings, ", ")
		}

		table.Append([]string{
			id[:8],
			fmt.Sprintf("%v", inst["name"]),
			fmt.Sprintf("%v", inst["image"]),
			fmt.Sprintf("%v", inst["status"]),
			access,
		})
	}
	table.Render()
}

func launchInstance() {
	qs := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Instance Name:",
			},
			Validate: survey.Required,
		},
		{
			Name: "image",
			Prompt: &survey.Select{
				Message: "Choose Image:",
				Options: []string{"alpine", "nginx:alpine", "ubuntu", "redis:alpine"},
				Default: "alpine",
			},
		},
		{
			Name: "ports",
			Prompt: &survey.Input{
				Message: "Port Mappings (host:container, optional):",
				Help:    "e.g. 8080:80",
			},
		},
	}

	answers := struct {
		Name  string
		Image string
		Ports string
	}{}

	if err := survey.Ask(qs, &answers); err != nil {
		return
	}

	client := getClient()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"name":  answers.Name,
			"image": answers.Image,
			"ports": answers.Ports,
		}).
		Post(apiURL + "/instances")

	if err != nil || resp.IsError() {
		fmt.Printf("❌ Failed: %v %s\n", err, resp.String())
		return
	}

	fmt.Printf("✅ Launched %s (%s) successfully!\n", answers.Name, answers.Image)
}

func stopInstance() {
	// First fetch list to select from
	client := getClient()
	resp, err := client.R().Get(apiURL + "/instances")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	json.Unmarshal(resp.Body(), &result)

	var options []string
	idMap := make(map[string]string) // "Display Name" -> UUID

	for _, inst := range result.Data {
		if inst["status"] != "RUNNING" {
			continue
		}
		id := fmt.Sprintf("%v", inst["id"])
		name := fmt.Sprintf("%v", inst["name"])
		display := fmt.Sprintf("%s (%s)", name, id[:8])
		options = append(options, display)
		idMap[display] = id
	}

	if len(options) == 0 {
		fmt.Println("⚠️  No running instances to stop.")
		return
	}

	var selected string
	prompt := &survey.Select{
		Message: "Select instance to stop:",
		Options: options,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return
	}

	targetID := idMap[selected]
	resp, err = client.R().Post(apiURL + "/instances/" + targetID + "/stop")
	if err != nil || resp.IsError() {
		fmt.Printf("❌ Failed to stop.\n")
		return
	}

	fmt.Printf("🛑 Stopping %s...\n", selected)
}

func viewLogs() {
	client := getClient()
	resp, err := client.R().Get(apiURL + "/instances")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	json.Unmarshal(resp.Body(), &result)

	var options []string
	idMap := make(map[string]string)

	for _, inst := range result.Data {
		id := fmt.Sprintf("%v", inst["id"])
		name := fmt.Sprintf("%v", inst["name"])
		display := fmt.Sprintf("%s (%s)", name, id[:8])
		options = append(options, display)
		idMap[display] = id
	}

	if len(options) == 0 {
		fmt.Println("⚠️  No instances found.")
		return
	}

	var selected string
	prompt := &survey.Select{
		Message: "Select instance to view logs:",
		Options: options,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return
	}

	targetID := idMap[selected]
	resp, err = client.R().Get(apiURL + "/instances/" + targetID + "/logs")
	if err != nil || resp.IsError() {
		fmt.Printf("❌ Failed to fetch logs.\n")
		return
	}

	fmt.Println("📜 --- Logs Start ---")
	fmt.Print(string(resp.Body()))
	fmt.Println("📜 --- Logs End ---")
}

func showInstance() {
	client := getClient()
	resp, err := client.R().Get(apiURL + "/instances")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	json.Unmarshal(resp.Body(), &result)

	var options []string
	idMap := make(map[string]string)

	for _, inst := range result.Data {
		id := fmt.Sprintf("%v", inst["id"])
		name := fmt.Sprintf("%v", inst["name"])
		display := fmt.Sprintf("%s (%s)", name, id[:8])
		options = append(options, display)
		idMap[display] = id
	}

	if len(options) == 0 {
		fmt.Println("⚠️  No instances found.")
		return
	}

	var selected string
	prompt := &survey.Select{
		Message: "Select instance to view details:",
		Options: options,
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return
	}

	targetID := idMap[selected]
	resp, err = client.R().Get(apiURL + "/instances/" + targetID)
	if err != nil || resp.IsError() {
		fmt.Printf("❌ Failed to fetch details.\n")
		return
	}

	var detailResult struct {
		Data map[string]interface{} `json:"data"`
	}
	json.Unmarshal(resp.Body(), &detailResult)
	inst := detailResult.Data

	fmt.Printf("\n☁️  Instance Details\n")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("%-15s %v\n", "ID:", inst["id"])
	fmt.Printf("%-15s %v\n", "Name:", inst["name"])
	fmt.Printf("%-15s %v\n", "Status:", inst["status"])
	fmt.Printf("%-15s %v\n", "Image:", inst["image"])
	fmt.Printf("%-15s %v\n", "Ports:", inst["ports"])
	fmt.Printf("%-15s %v\n", "Created At:", inst["created_at"])
	fmt.Printf("%-15s %v\n", "Version:", inst["version"])
	fmt.Printf("%-15s %v\n", "Container ID:", inst["container_id"])
	fmt.Println(strings.Repeat("-", 40))
}
