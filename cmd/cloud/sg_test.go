package main

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestSGCommands(t *testing.T) {
	if sgCmd == nil {
		t.Fatal("sgCmd is nil")
	}
	if !sgCmd.HasSubCommands() {
		t.Fatal("sgCmd should have subcommands")
	}

	subs := map[string][]string{
		"create":      {"vpc-id", "description"},
		"list":        {"vpc-id"},
		"get":         {},
		"delete":      {},
		"add-rule":    {"direction", "protocol", "port-min", "port-max", "cidr", "priority"},
		"remove-rule": {},
		"attach":      {},
		"detach":      {},
	}

	for cmdName, flags := range subs {
		var sub *cobra.Command
		for _, c := range sgCmd.Commands() {
			if c.Name() == cmdName {
				sub = c
				break
			}
		}
		if sub == nil {
			t.Errorf("sgCmd should have subcommand %q", cmdName)
			continue
		}

		for _, flag := range flags {
			if sub.Flag(flag) == nil {
				t.Errorf("subcommand %q should have flag %q", cmdName, flag)
			}
		}
	}
}
