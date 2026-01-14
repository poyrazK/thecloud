// Package main provides a libvirt connectivity test harness.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/poyrazk/thecloud/internal/core/ports"
	"github.com/poyrazk/thecloud/internal/repositories/libvirt"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	adapter, err := libvirt.NewLibvirtAdapter(logger, "/var/run/libvirt/libvirt-sock")
	if err != nil {
		fmt.Printf("Failed to connect to libvirt: %v\n", err)
		fmt.Println("Make sure libvirtd is running and you have access to /var/run/libvirt/libvirt-sock")
		return
	}

	fmt.Println("--- Testing Libvirt Adapter ---")

	// 1. Ping
	if err := adapter.Ping(ctx); err != nil {
		fmt.Printf("Ping failed: %v\n", err)
		return
	}
	fmt.Println("[✓] Ping successful")

	// 2. Use default network (skip custom network creation for simplicity)
	netName := "default"
	fmt.Println("[✓] Using default network")

	// 3. Create Instance with Cloud-Init
	fmt.Println("Creating test instance (this requires genisoimage on host)...")
	instanceName := "test-vm-manual"
	env := []string{"TEST_VAR=hello_world"}
	cmd := []string{"echo 'Cloud-Init Worked' > /tmp/success"}

	// We use 'alpine' as a dummy image name.
	// NOTE: This test expects an 'alpine-root' or similar to be setup if we had a real pool.
	// Since we implement a "create empty volume" fallback in CreateInstance, it should at least define the VM.
	id, err := adapter.CreateInstance(ctx, ports.CreateInstanceOptions{
		Name:      instanceName,
		ImageName: "alpine",
		Ports:     []string{"8080:80"},
		NetworkID: netName,
		Env:       env,
		Cmd:       cmd,
	})
	if err != nil {
		fmt.Printf("CreateInstance failed: %v\n", err)
		return
	}
	fmt.Printf("[✓] Instance created: %s\n", id)
	defer func() {
		if err := adapter.DeleteInstance(ctx, id); err != nil {
			fmt.Printf("Warning: cleanup failed: %v\n", err)
		}
	}()

	// 4. Check IP (Wait a bit for DHCP)
	fmt.Println("Waiting for IP (this might fail if VM doesn't actually boot a real OS)...")
	ipCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for {
		select {
		case <-ipCtx.Done():
			fmt.Println("[!] Timeout waiting for IP")
			goto next
		case <-time.After(5 * time.Second):
			ip, err := adapter.GetInstanceIP(ctx, id)
			if err == nil && ip != "" {
				fmt.Printf("[✓] Instance IP detected: %s\n", ip)
				goto next
			}
		}
	}

next:
	// 5. Cleanup is handled by defers
	fmt.Println("--- Test Complete (Cleaning up) ---")
}
