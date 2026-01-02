package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func printJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func printError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

func printStatus(message string) {
	if outputJSON {
		printJSON(map[string]string{
			"status":  "ok",
			"message": message,
		})
		return
	}
	fmt.Println(message)
}

func printDataOrStatus(data any, message string) {
	if outputJSON {
		printJSON(data)
		return
	}
	if message != "" {
		fmt.Println(message)
	}
}
