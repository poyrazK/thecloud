package sshutil

import "fmt"

func ExampleClient() {
	client := &Client{Host: "127.0.0.1:22", User: "user"}
	fmt.Println(client.Host)
	// Output: 127.0.0.1:22
}
