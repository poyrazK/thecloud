// Package libvirt provides load balancer proxy implementation using host-level Nginx.
package libvirt

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/domain"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

const (
	nginxConfigFileName = "nginx.conf"
	nginxPidFileName    = "nginx.pid"
)

// LBProxyAdapter manages host-level Nginx load balancer proxies.
type LBProxyAdapter struct {
	compute ports.ComputeBackend
}

// NewLBProxyAdapter constructs an LBProxyAdapter for the given compute backend.
func NewLBProxyAdapter(compute ports.ComputeBackend) *LBProxyAdapter {
	return &LBProxyAdapter{
		compute: compute,
	}
}

func (a *LBProxyAdapter) DeployProxy(ctx context.Context, lb *domain.LoadBalancer, targets []*domain.LBTarget) (string, error) {
	config, err := a.generateNginxConfig(ctx, lb, targets)
	if err != nil {
		return "", err
	}

	configDir := filepath.Join("/tmp", "thecloud", "lb", lb.ID.String())
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	configPath := filepath.Join(configDir, nginxConfigFileName)
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return "", err
	}

	pidPath := filepath.Join(configDir, nginxPidFileName)
	// Start nginx
	// nginx -c /path/to/conf -g "pid /path/to/pid; daemon on;"
	cmd := execCommandContext(ctx, "nginx", "-c", configPath, "-g", fmt.Sprintf("pid %s; daemon on;", pidPath))
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to start nginx: %w", err)
	}

	return lb.ID.String(), nil
}

func (a *LBProxyAdapter) RemoveProxy(ctx context.Context, lbID uuid.UUID) error {
	configDir := filepath.Join("/tmp", "thecloud", "lb", lbID.String())
	pidPath := filepath.Join(configDir, nginxPidFileName)

	if _, err := os.Stat(pidPath); err == nil {
		// Stop nginx: nginx -s stop -c ...
		configPath := filepath.Join(configDir, nginxConfigFileName)
		cmd := execCommandContext(ctx, "nginx", "-c", configPath, "-g", fmt.Sprintf("pid %s;", pidPath), "-s", "stop")
		_ = cmd.Run()
	}

	_ = os.RemoveAll(configDir)
	return nil
}

func (a *LBProxyAdapter) UpdateProxyConfig(ctx context.Context, lb *domain.LoadBalancer, targets []*domain.LBTarget) error {
	config, err := a.generateNginxConfig(ctx, lb, targets)
	if err != nil {
		return err
	}

	configDir := filepath.Join("/tmp", "thecloud", "lb", lb.ID.String())
	configPath := filepath.Join(configDir, nginxConfigFileName)
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return err
	}

	pidPath := filepath.Join(configDir, nginxPidFileName)
	// Reload nginx
	cmd := execCommandContext(ctx, "nginx", "-c", configPath, "-g", fmt.Sprintf("pid %s;", pidPath), "-s", "reload")
	return cmd.Run()
}

func (a *LBProxyAdapter) generateNginxConfig(ctx context.Context, lb *domain.LoadBalancer, targets []*domain.LBTarget) (string, error) {
	tmplRaw := `
user root;
events {
    worker_connections 1024;
}

http {
    {{if .Targets}}
    upstream backend {
        {{range .Targets}}
        server {{.IP}}:{{.Port}} weight={{.Weight}};
        {{end}}
        {{if .LeastConn}}least_conn;{{end}}
    }
    {{end}}

    server {
        listen {{.Port}};
        location / {
            {{if .Targets}}
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            {{else}}
            return 503 "No targets available";
            {{end}}
        }
    }
}
`
	type targetInfo struct {
		IP     string
		Port   int
		Weight int
	}
	type data struct {
		Port      int
		LeastConn bool
		Targets   []targetInfo
	}

	d := data{
		Port:      lb.Port,
		LeastConn: lb.Algorithm == "least-conn",
	}

	for _, t := range targets {
		ip, err := a.compute.GetInstanceIP(ctx, t.InstanceID.String())
		if err != nil {
			continue
		}
		d.Targets = append(d.Targets, targetInfo{
			IP:     ip,
			Port:   t.Port,
			Weight: t.Weight,
		})
	}

	tmpl, err := template.New("nginx").Parse(tmplRaw)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, d); err != nil {
		return "", err
	}

	return buf.String(), nil
}
