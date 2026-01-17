// Package libvirt provides a Libvirt/KVM implementation of the ComputeBackend interface.
// It enables running VMs using QEMU/KVM with features like Cloud-Init, volume management,
// snapshots, and network configuration.
package libvirt

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/google/uuid"
	"github.com/poyrazk/thecloud/internal/core/ports"
)

var execCommand = exec.Command
var execCommandContext = exec.CommandContext
var mkdirTemp = os.MkdirTemp
var lookPath = exec.LookPath
var osOpen = os.Open

const (
	defaultPoolName   = "default"
	userDataFileName  = "user-data"
	metaDataFileName  = "meta-data"
	errGetVolumePath  = "failed to get volume path: %w"
	errDomainNotFound = "domain not found: %w"
	errPoolNotFound   = "storage pool not found: %w"
	prefixCloudInit   = "cloud-init-"
)

// LibvirtAdapter implements compute backend operations using libvirt/KVM.
type LibvirtAdapter struct {
	client LibvirtClient
	logger *slog.Logger
	uri    string

	mu           sync.RWMutex
	portMappings map[string]map[string]int // instanceID -> internalPort -> hostPort

	// Network pool configuration
	networkCounter   int
	poolStart        net.IP
	poolEnd          net.IP
	ipWaitInterval   time.Duration
	taskWaitInterval time.Duration
}

// NewLibvirtAdapter creates a LibvirtAdapter connected to the provided URI.
func NewLibvirtAdapter(logger *slog.Logger, uri string) (*LibvirtAdapter, error) {
	if uri == "" {
		uri = "/var/run/libvirt/libvirt-sock"
	}

	// Connect to libvirt
	// We use a dialer for the unix socket
	c, err := net.DialTimeout("unix", uri, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to dial libvirt: %w", err)
	}

	//nolint:staticcheck // libvirt.New is deprecated but NewWithDialer doesn't work with our setup
	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}

	return &LibvirtAdapter{
		client:         &RealLibvirtClient{conn: l},
		logger:         logger,
		uri:            uri,
		portMappings:   make(map[string]map[string]int),
		networkCounter: 0,
		// Local private network range for VM pools - safe for internal usage
		poolStart:        net.ParseIP("192.168.100.0"),
		poolEnd:          net.ParseIP("192.168.200.255"),
		ipWaitInterval:   5 * time.Second,
		taskWaitInterval: 2 * time.Second,
	}, nil
}

// Close gracefully disconnects from libvirt
func (a *LibvirtAdapter) Close() error {
	return a.client.Close()
}

// Ping checks if libvirt is reachable
func (a *LibvirtAdapter) Ping(ctx context.Context) error {
	return a.client.Connect(ctx) // returns error only
}

func (a *LibvirtAdapter) Type() string {
	return "libvirt"
}

func (a *LibvirtAdapter) CreateInstance(ctx context.Context, opts ports.CreateInstanceOptions) (string, error) {
	name := a.sanitizeDomainName(opts.Name)

	diskPath, vol, err := a.prepareRootVolume(ctx, name)
	if err != nil {
		return "", err
	}

	isoPath := a.prepareCloudInit(ctx, name, opts.Env, opts.Cmd)
	additionalDisks := a.resolveBinds(ctx, opts.VolumeBinds)

	networkID := opts.NetworkID
	if networkID == "" {
		networkID = "default"
	}

	domainXML := generateDomainXML(name, diskPath, networkID, isoPath, 512, 1, additionalDisks)
	dom, err := a.client.DomainDefineXML(ctx, domainXML)
	if err != nil {
		a.cleanupCreateFailure(ctx, vol, isoPath)
		return "", fmt.Errorf("failed to define domain: %w", err)
	}

	if err := a.client.DomainCreate(ctx, dom); err != nil {
		_ = a.client.DomainUndefine(ctx, dom)
		_ = a.client.StorageVolDelete(ctx, vol, 0)
		return "", fmt.Errorf("failed to start domain: %w", err)
	}

	if len(opts.Ports) > 0 {
		go a.setupPortForwarding(name, opts.Ports)
	}

	return name, nil
}

func (a *LibvirtAdapter) sanitizeDomainName(name string) string {
	name = regexp.MustCompile(`[^a-zA-Z0-9-]`).ReplaceAllString(name, "")
	if name == "" {
		return uuid.New().String()[:8]
	}
	return name
}

func (a *LibvirtAdapter) prepareCloudInit(ctx context.Context, name string, env []string, cmd []string) string {
	if len(env) == 0 && len(cmd) == 0 {
		return ""
	}
	isoPath, err := a.generateCloudInitISO(ctx, name, env, cmd)
	if err != nil {
		a.logger.Warn("failed to generate cloud-init iso, proceeding without it", "error", err)
		return ""
	}
	return isoPath
}

func (a *LibvirtAdapter) cleanupCreateFailure(ctx context.Context, vol libvirt.StorageVol, isoPath string) {
	_ = a.client.StorageVolDelete(ctx, vol, 0)
	if isoPath != "" {
		if err := os.Remove(isoPath); err != nil {
			a.logger.Warn("failed to remove ISO", "path", isoPath, "error", err)
		}
	}
}

func (a *LibvirtAdapter) waitInitialIP(ctx context.Context, id string) (string, error) {
	ticker := time.NewTicker(a.ipWaitInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
			ip, err := a.GetInstanceIP(ctx, id)
			if err == nil && ip != "" {
				return ip, nil
			}
		}
	}
}

func (a *LibvirtAdapter) StopInstance(ctx context.Context, id string) error {
	dom, err := a.client.DomainLookupByName(ctx, id)
	if err != nil {
		return fmt.Errorf(errDomainNotFound, err)
	}

	if err := a.client.DomainDestroy(ctx, dom); err != nil {
		return fmt.Errorf("failed to destroy domain: %w", err)
	}
	return nil
}

func (a *LibvirtAdapter) DeleteInstance(ctx context.Context, id string) error {
	dom, err := a.client.DomainLookupByName(ctx, id)
	if err != nil {
		return nil
	}

	a.stopDomainIfRunning(ctx, dom)

	if err := a.client.DomainUndefine(ctx, dom); err != nil {
		return fmt.Errorf("failed to undefine domain: %w", err)
	}

	a.cleanupPortMappings(id)
	a.cleanupDomainISO(id)
	a.cleanupRootVolume(ctx, id)

	return nil
}

func (a *LibvirtAdapter) stopDomainIfRunning(ctx context.Context, dom libvirt.Domain) {
	state, _, err := a.client.DomainGetState(ctx, dom, 0)
	if err == nil && state == 1 { // Running
		_ = a.client.DomainDestroy(ctx, dom)
	}
}

func (a *LibvirtAdapter) cleanupPortMappings(id string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if mappings, ok := a.portMappings[id]; ok {
		for _, hPort := range mappings {
			_ = execCommand("sudo", "iptables", "-t", "nat", "-D", "PREROUTING", "-p", "tcp", "--dport", fmt.Sprintf("%d", hPort), "-j", "DNAT").Run()
		}
		delete(a.portMappings, id)
	}
}

func (a *LibvirtAdapter) cleanupDomainISO(id string) {
	if err := validateID(id); err != nil {
		a.logger.Warn("invalid id for iso cleanup", "id", id)
		return
	}
	isoPath := filepath.Join(os.TempDir(), prefixCloudInit+id+".iso")
	_ = os.Remove(isoPath)
}

func (a *LibvirtAdapter) cleanupRootVolume(ctx context.Context, id string) {
	pool, err := a.client.StoragePoolLookupByName(ctx, defaultPoolName)
	if err != nil {
		return
	}
	volName := id + "-root"
	vol, err := a.client.StorageVolLookupByName(ctx, pool, volName)
	if err == nil {
		_ = a.client.StorageVolDelete(ctx, vol, 0)
	}
}

func (a *LibvirtAdapter) GetInstanceLogs(ctx context.Context, id string) (io.ReadCloser, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	// Read from standard qemu log location
	// Note: This contains QEMU output, not necessarily guest console output unless serial is redirected there.
	// To get guest console, we'd need to attach to console or read a file if defined in XML.
	// Our XML defined <console type='pty'> which goes to a PTY. Reading PTY from outside is complex.
	// We'll fall back to QEMU log for debug info.
	logPath := fmt.Sprintf("/var/log/libvirt/qemu/%s.log", id)
	f, err := osOpen(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	return f, nil
}

func (a *LibvirtAdapter) GetInstanceStats(ctx context.Context, id string) (io.ReadCloser, error) {
	dom, err := a.client.DomainLookupByName(ctx, id)
	if err != nil {
		return nil, fmt.Errorf(errDomainNotFound, err)
	}

	// Memory stats
	// We use standard libvirt stats. The format for ComputeBackend expects JSON.
	// We'll construct a simple JSON.
	memStats, err := a.client.DomainMemoryStats(ctx, dom, 10, 0)
	if err != nil {
		return nil, err
	}

	var usage, limit uint64
	for _, stat := range memStats {
		if stat.Tag == 6 { // rss
			usage = stat.Val * 1024 // KB to Bytes
		}
		if stat.Tag == 5 { // actual
			limit = stat.Val * 1024
		}
	}

	json := fmt.Sprintf(`{"memory_stats":{"usage":%d,"limit":%d}}`, usage, limit)
	return io.NopCloser(strings.NewReader(json)), nil
}

func (a *LibvirtAdapter) GetInstancePort(ctx context.Context, id string, internalPort string) (int, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if mappings, ok := a.portMappings[id]; ok {
		if hostPort, ok := mappings[internalPort]; ok {
			return hostPort, nil
		}
	}

	return 0, fmt.Errorf("no host port mapping found for instance %s port %s", id, internalPort)
}

func (a *LibvirtAdapter) AttachVolume(ctx context.Context, id string, volumePath string) error {
	dom, err := a.client.DomainLookupByName(ctx, id)
	if err != nil {
		return fmt.Errorf(errDomainNotFound, err)
	}

	// We need to find an available device name (vdb, vdc, etc.)
	// For simplicity, we'll try to get XML and check existing disks.
	// In a real implementation we'd have a counter or better logic.
	dev := "vdb" // Hardcoded for POC

	diskType := "file"
	driverType := "qcow2"
	sourceAttr := "file"

	if strings.HasPrefix(volumePath, "/dev/") {
		diskType = "block"
		driverType = "raw"
		sourceAttr = "dev"
	}

	xml := fmt.Sprintf(`
    <disk type='%s' device='disk'>
      <driver name='qemu' type='%s'/>
      <source %s='%s'/>
      <target dev='%s' bus='virtio'/>
    </disk>`, diskType, driverType, sourceAttr, volumePath, dev)

	return a.client.DomainAttachDevice(ctx, dom, xml)
}

func (a *LibvirtAdapter) DetachVolume(ctx context.Context, id string, volumePath string) error {
	dom, err := a.client.DomainLookupByName(ctx, id)
	if err != nil {
		return fmt.Errorf(errDomainNotFound, err)
	}

	// To detach, we technically only need the target device or a matching XML.
	// For simplicity, we construct a matching XML.
	// Note: We'd need the same XML or at least target/source.
	xml := fmt.Sprintf("<disk type='file' device='disk'><source file='%s'/><target dev='vdb' bus='virtio'/></disk>", volumePath)
	if strings.HasPrefix(volumePath, "/dev/") {
		xml = fmt.Sprintf("<disk type='block' device='disk'><source dev='%s'/><target dev='vdb' bus='virtio'/></disk>", volumePath)
	}

	return a.client.DomainDetachDevice(ctx, dom, xml)
}

func (a *LibvirtAdapter) GetConsoleURL(ctx context.Context, id string) (string, error) {
	if err := validateID(id); err != nil {
		return "", err
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	domain, err := a.client.DomainLookupByName(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to lookup domain: %w", err)
	}

	xml, err := a.client.DomainGetXMLDesc(ctx, domain, 0)
	if err != nil {
		return "", fmt.Errorf("failed to get domain xml: %w", err)
	}

	// Simple string parsing for VNC port as we don't want to import heavy XML decoders if not needed
	// XML looks like: <graphics type='vnc' port='5900' ... />
	startIdx := strings.Index(xml, "<graphics type='vnc' port='")
	if startIdx == -1 {
		return "", fmt.Errorf("vnc graphics not found in domain xml")
	}
	xml = xml[startIdx+len("<graphics type='vnc' port='"):]
	endIdx := strings.Index(xml, "'")
	if endIdx == -1 {
		return "", fmt.Errorf("failed to parse vnc port from xml")
	}

	portStr := xml[:endIdx]
	return fmt.Sprintf("vnc://%s:%s", "127.0.0.1", portStr), nil
}

func (a *LibvirtAdapter) GetInstanceIP(ctx context.Context, id string) (string, error) {
	// 1. Get Domain
	dom, err := a.client.DomainLookupByName(ctx, id)
	if err != nil {
		return "", fmt.Errorf(errDomainNotFound, err)
	}

	// 2. We need the MAC address to look up DHCP leases.
	// We can get XML desc and parse it.
	xmlDesc, err := a.client.DomainGetXMLDesc(ctx, dom, 0)
	if err != nil {
		return "", fmt.Errorf("failed to get domain xml: %w", err)
	}

	// Extract MAC using string parsing (dirty but lighter than XML decoder for one field)
	// Look for <mac address='XX:XX:XX:XX:XX:XX'/>
	// This assumes one interface.
	start := strings.Index(xmlDesc, "<mac address='")
	if start == -1 {
		return "", fmt.Errorf("no mac address found in xml")
	}
	start += len("<mac address='")
	end := strings.Index(xmlDesc[start:], "'/>")
	if end == -1 {
		return "", fmt.Errorf("malformed xml mac address")
	}
	mac := xmlDesc[start : start+end]

	// 3. Lookup leases in default network
	// We assume "default" network
	net, err := a.client.NetworkLookupByName(ctx, "default")
	if err != nil {
		return "", fmt.Errorf("default network not found: %w", err)
	}

	// Pass nil for mac to get all leases (simplifies type handling of OptString)
	leases, _, err := a.client.NetworkGetDhcpLeases(ctx, net, nil, 0, 0)
	if err != nil {
		return "", fmt.Errorf("failed to get leases: %w", err)
	}

	for _, lease := range leases {
		if len(lease.Mac) > 0 && lease.Mac[0] == mac {
			return lease.Ipaddr, nil
		}
	}

	return "", fmt.Errorf("no ip lease found for %s (%s)", id, mac)
}

func (a *LibvirtAdapter) Exec(ctx context.Context, id string, cmd []string) (string, error) {
	return "", fmt.Errorf("exec not supported in libvirt adapter")
}

func (a *LibvirtAdapter) RunTask(ctx context.Context, opts ports.RunTaskOptions) (string, error) {
	name := "task-" + uuid.New().String()[:8]

	pool, err := a.client.StoragePoolLookupByName(ctx, defaultPoolName)
	if err != nil {
		return "", fmt.Errorf("failed to find default pool: %w", err)
	}

	volXML := generateVolumeXML(name+"-root", 1)
	vol, err := a.client.StorageVolCreateXML(ctx, pool, volXML, 0)
	if err != nil {
		return "", fmt.Errorf("failed to create root volume: %w", err)
	}

	diskPath, err := a.client.StorageVolGetPath(ctx, vol)
	if err != nil {
		_ = a.client.StorageVolDelete(ctx, vol, 0)
		return "", fmt.Errorf(errGetVolumePath, err)
	}

	isoPath := a.prepareCloudInit(ctx, name, nil, opts.Command)
	additionalDisks := a.resolveBinds(ctx, opts.Binds)
	domainXML := generateDomainXML(name, diskPath, "default", isoPath, int(opts.MemoryMB), 1, additionalDisks)

	dom, err := a.client.DomainDefineXML(ctx, domainXML)
	if err != nil {
		a.cleanupCreateFailure(ctx, vol, isoPath)
		return "", fmt.Errorf("failed to define domain: %w", err)
	}

	if err := a.client.DomainCreate(ctx, dom); err != nil {
		_ = a.client.DomainUndefine(ctx, dom)
		_ = a.client.StorageVolDelete(ctx, vol, 0)
		return "", fmt.Errorf("failed to start domain: %w", err)
	}

	return name, nil
}

func (a *LibvirtAdapter) WaitTask(ctx context.Context, id string) (int64, error) {
	// Poll for domain state to be Shutoff
	// Since we can't easily get the exit code from inside the VM without qemu-agent,
	// we assume 0 if it shuts down gracefully (state Shutoff).

	interval := a.taskWaitInterval
	if interval == 0 {
		interval = 2 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return -1, ctx.Err()
		case <-ticker.C:
			dom, err := a.client.DomainLookupByName(ctx, id)
			if err != nil {
				// If domain is gone, maybe it was deleted?
				return -1, fmt.Errorf(errDomainNotFound, err)
			}

			state, _, err := a.client.DomainGetState(ctx, dom, 0)
			if err != nil {
				continue
			}

			// libvirt.DomainShutoff = 5
			if state == 5 {
				return 0, nil
			}
		}
	}
}

func (a *LibvirtAdapter) CreateNetwork(ctx context.Context, name string) (string, error) {
	// Allocate network range from pool
	gateway, rangeStart, rangeEnd := a.getNextNetworkRange()

	// Simple NAT network
	xml := generateNetworkXML(name, "virbr-"+name, gateway, rangeStart, rangeEnd)

	net, err := a.client.NetworkDefineXML(ctx, xml)
	if err != nil {
		return "", fmt.Errorf("failed to define network: %w", err)
	}

	if err := a.client.NetworkCreate(ctx, net); err != nil {
		return "", fmt.Errorf("failed to start network: %w", err)
	}

	return net.Name, nil
}

func (a *LibvirtAdapter) DeleteNetwork(ctx context.Context, id string) error {
	net, err := a.client.NetworkLookupByName(ctx, id)
	if err != nil {
		return nil // assume deleted
	}

	if err := a.client.NetworkDestroy(ctx, net); err != nil {
		a.logger.Warn("failed to destroy network", "error", err)
	}
	if err := a.client.NetworkUndefine(ctx, net); err != nil {
		return fmt.Errorf("failed to undefine network: %w", err)
	}
	return nil
}

func (a *LibvirtAdapter) CreateVolume(ctx context.Context, name string) error {
	pool, err := a.client.StoragePoolLookupByName(ctx, defaultPoolName)
	if err != nil {
		return fmt.Errorf(errPoolNotFound, err)
	}

	// 10GB default
	xml := generateVolumeXML(name, 10)

	// Refresh pool first
	if err := a.client.StoragePoolRefresh(ctx, pool, 0); err != nil {
		// Log but continue
		a.logger.Warn("failed to refresh pool", "error", err)
	}

	_, err = a.client.StorageVolCreateXML(ctx, pool, xml, 0)
	if err != nil {
		return fmt.Errorf("failed to create volume xml: %w", err)
	}
	return nil
}

func (a *LibvirtAdapter) DeleteVolume(ctx context.Context, name string) error {
	pool, err := a.client.StoragePoolLookupByName(ctx, defaultPoolName)
	if err != nil {
		return fmt.Errorf(errPoolNotFound, err)
	}

	vol, err := a.client.StorageVolLookupByName(ctx, pool, name)
	if err != nil {
		// Check if not found
		return nil
	}

	if err := a.client.StorageVolDelete(ctx, vol, 0); err != nil {
		return fmt.Errorf("failed to delete volume: %w", err)
	}
	return nil
}

func (a *LibvirtAdapter) CreateVolumeSnapshot(ctx context.Context, volumeID string, destinationPath string) error {
	// volumeID is the libvirt volume name
	pool, err := a.client.StoragePoolLookupByName(ctx, defaultPoolName)
	if err != nil {
		return fmt.Errorf(errPoolNotFound, err)
	}

	vol, err := a.client.StorageVolLookupByName(ctx, pool, volumeID)
	if err != nil {
		return fmt.Errorf("failed to find volume: %w", err)
	}

	volPath, err := a.client.StorageVolGetPath(ctx, vol)
	if err != nil {
		return fmt.Errorf(errGetVolumePath, err)
	}

	// Use qemu-img to convert the volume to a temporary qcow2
	tmpQcow2 := destinationPath + ".qcow2"
	cmd := execCommand("qemu-img", "convert", "-O", "qcow2", volPath, tmpQcow2)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("qemu-img convert failed: %w", err)
	}
	defer func() {
		if err := os.Remove(tmpQcow2); err != nil {
			a.logger.Warn("failed to remove temp file", "path", tmpQcow2, "error", err)
		}
	}()

	// Now tar it to match the SnapshotService expectation (tarball of contents)
	// Actually, if we just want a "blob", a tarred qcow2 is fine.
	tarCmd := execCommand("tar", "czf", destinationPath, "-C", filepath.Dir(tmpQcow2), filepath.Base(tmpQcow2))
	if err := tarCmd.Run(); err != nil {
		return fmt.Errorf("tar archive failed: %w", err)
	}

	return nil
}

func (a *LibvirtAdapter) RestoreVolumeSnapshot(ctx context.Context, volumeID string, sourcePath string) error {
	pool, err := a.client.StoragePoolLookupByName(ctx, defaultPoolName)
	if err != nil {
		return fmt.Errorf(errPoolNotFound, err)
	}

	vol, err := a.client.StorageVolLookupByName(ctx, pool, volumeID)
	if err != nil {
		return fmt.Errorf("failed to find volume: %w", err)
	}

	volPath, err := a.client.StorageVolGetPath(ctx, vol)
	if err != nil {
		return fmt.Errorf(errGetVolumePath, err)
	}

	// 1. Untar
	tmpDir, err := mkdirTemp("", "restore-")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			a.logger.Warn("failed to remove temp dir", "path", tmpDir, "error", err)
		}
	}()

	untarCmd := execCommand("tar", "xzf", sourcePath, "-C", tmpDir)
	if err := untarCmd.Run(); err != nil {
		return fmt.Errorf("untar failed: %w", err)
	}

	files, _ := os.ReadDir(tmpDir)
	if len(files) == 0 {
		return fmt.Errorf("empty snapshot archive")
	}
	tmpQcow2 := filepath.Join(tmpDir, files[0].Name())

	// 2. Restore using qemu-img
	cmd := execCommand("qemu-img", "convert", "-O", "qcow2", tmpQcow2, volPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("qemu-img restore failed: %w", err)
	}

	return nil
}
func (a *LibvirtAdapter) generateCloudInitISO(_ context.Context, name string, env []string, cmd []string) (string, error) {
	safeName := a.sanitizeDomainName(name)

	// Additional security: ensure the sanitized name doesn't contain path separators
	safeName = filepath.Base(safeName)
	if safeName == "." || safeName == ".." {
		safeName = fmt.Sprintf("cloudinit-%d", time.Now().UnixNano())
	}

	tmpDir, err := os.MkdirTemp("", prefixCloudInit+safeName)
	if err != nil {
		return "", err
	}
	defer a.cleanupTempDir(tmpDir)

	if err := a.writeCloudInitFiles(tmpDir, name, env, cmd); err != nil {
		return "", err
	}

	// Use filepath.Join to safely construct the ISO path
	isoPath := filepath.Join(os.TempDir(), filepath.Clean(prefixCloudInit+safeName+".iso"))
	return isoPath, a.runIsoCommand(isoPath, tmpDir)
}

func (a *LibvirtAdapter) cleanupTempDir(tmpDir string) {
	if err := os.RemoveAll(tmpDir); err != nil {
		a.logger.Warn("failed to remove temp dir", "path", tmpDir, "error", err)
	}
}

func (a *LibvirtAdapter) writeCloudInitFiles(tmpDir, name string, env, cmd []string) error {
	metaData := fmt.Sprintf("instance-id: %s\nlocal-hostname: %s\n", name, name)
	if err := os.WriteFile(filepath.Join(tmpDir, metaDataFileName), []byte(metaData), 0644); err != nil {
		return err
	}

	userData := a.generateUserData(env, cmd)
	return os.WriteFile(filepath.Join(tmpDir, userDataFileName), userData, 0644)
}

func (a *LibvirtAdapter) generateUserData(env, cmd []string) []byte {
	var userData bytes.Buffer
	userData.WriteString("#cloud-config\n")

	if len(env) > 0 {
		userData.WriteString("write_files:\n")
		userData.WriteString("  - path: /etc/profile.d/cloud-env.sh\n")
		userData.WriteString("    content: |\n")
		for _, e := range env {
			userData.WriteString(fmt.Sprintf("      export %s\n", e))
		}
	}

	if len(cmd) > 0 {
		userData.WriteString("runcmd:\n")
		for _, c := range cmd {
			userData.WriteString(fmt.Sprintf("  - [ sh, -c, %q ]\n", c))
		}
	}
	return userData.Bytes()
}

func (a *LibvirtAdapter) runIsoCommand(isoPath, tmpDir string) error {
	genCmd := execCommand("genisoimage", "-output", isoPath, "-volid", "config-2", "-joliet", "-rock",
		filepath.Join(tmpDir, userDataFileName), filepath.Join(tmpDir, metaDataFileName))

	if _, err := genCmd.CombinedOutput(); err != nil {
		genCmd = execCommand("mkisofs", "-output", isoPath, "-volid", "config-2", "-joliet", "-rock",
			filepath.Join(tmpDir, userDataFileName), filepath.Join(tmpDir, metaDataFileName))
		if _, err2 := genCmd.CombinedOutput(); err2 != nil {
			return fmt.Errorf("failed to generate iso (genisoimage/mkisofs): %w", err2)
		}
	}
	return nil
}

// getNextNetworkRange allocates the next /24 network from the pool
func (a *LibvirtAdapter) getNextNetworkRange() (gateway, rangeStart, rangeEnd string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Calculate base IP: poolStart + (networkCounter * 256)
	baseIP := make(net.IP, len(a.poolStart))
	copy(baseIP, a.poolStart)

	// Add offset for this network (each network gets a /24)
	offset := a.networkCounter * 256
	for i := len(baseIP) - 1; i >= 0 && offset > 0; i-- {
		sum := int(baseIP[i]) + offset
		baseIP[i] = byte(sum % 256)
		offset = sum / 256
	}

	a.networkCounter++

	// Gateway is .1, DHCP range is .2 to .254
	gw := make(net.IP, len(baseIP))
	copy(gw, baseIP)
	gw[len(gw)-1] = 1

	start := make(net.IP, len(baseIP))
	copy(start, baseIP)
	start[len(start)-1] = 2

	end := make(net.IP, len(baseIP))
	copy(end, baseIP)
	end[len(end)-1] = 254

	return gw.String(), start.String(), end.String()
}

func validateID(id string) error {
	if strings.Contains(id, "..") || strings.Contains(id, "/") || strings.Contains(id, "\\") {
		return fmt.Errorf("invalid id: contains path traversal characters")
	}
	return nil
}

func (a *LibvirtAdapter) setupPortForwarding(name string, ports []string) {
	// Wait for VM to get an IP
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	ip, err := a.waitInitialIP(ctx, name)
	if err != nil {
		a.logger.Error("failed to get ip for port forwarding", "instance", name, "error", err)
		return
	}

	for _, p := range ports {
		hPort, cP, err := a.parseAndValidatePort(p)
		if err != nil {
			a.logger.Warn("skipping invalid port forwarding configuration", "port", p, "error", err)
			continue
		}

		// Validate IP format
		if net.ParseIP(ip) == nil {
			a.logger.Error("invalid vm ip for port forwarding", "ip", ip)
			return
		}

		if hPort > 0 {
			a.configureIptables(name, ip, strconv.Itoa(cP), hPort, cP)
		}
	}
}

func (a *LibvirtAdapter) configureIptables(name, ip, containerPort string, hPort, cP int) {
	a.mu.Lock()
	if a.portMappings[name] == nil {
		a.portMappings[name] = make(map[string]int)
	}
	a.portMappings[name][containerPort] = hPort
	a.mu.Unlock()

	a.logger.Info("setting up port forwarding", "host", hPort, "vm", containerPort, "ip", ip)
	// iptables -t nat -A PREROUTING -p tcp --dport <hPort> -j DNAT --to <ip>:<containerPort>
	path, err := lookPath("iptables")
	if err != nil {
		a.logger.Error("iptables not found, cannot set up port forwarding", "error", err)
		return
	}
	if path != "" {
		// Security: Use numeric values to avoid injection
		hPortStr := strconv.Itoa(hPort)
		cPortStr := strconv.Itoa(cP)
		cmd := execCommand("sudo", "iptables", "-t", "nat", "-A", "PREROUTING", "-p", "tcp", "--dport", hPortStr, "-j", "DNAT", "--to", ip+":"+cPortStr)
		if err := cmd.Run(); err != nil {
			a.logger.Error("failed to set up iptables rule", "command", cmd.String(), "error", err)
		}
	}
}

func (a *LibvirtAdapter) prepareRootVolume(ctx context.Context, name string) (string, libvirt.StorageVol, error) {
	// We assume 'imageName' is a backing volume in the default pool.
	pool, err := a.client.StoragePoolLookupByName(ctx, defaultPoolName)
	if err != nil {
		return "", libvirt.StorageVol{}, fmt.Errorf("default pool not found: %w", err)
	}

	// Create root disk for the VM
	// For simplicity, we just create an empty 10GB qcows if image not found, or clone if we knew how
	volXML := generateVolumeXML(name+"-root", 10)

	vol, err := a.client.StorageVolCreateXML(ctx, pool, volXML, 0)
	if err != nil {
		return "", libvirt.StorageVol{}, fmt.Errorf("failed to create root volume: %w", err)
	}

	// Get volume path from libvirt
	diskPath, err := a.client.StorageVolGetPath(ctx, vol)
	if err != nil {
		_ = a.client.StorageVolDelete(ctx, vol, 0)
		return "", libvirt.StorageVol{}, fmt.Errorf(errGetVolumePath, err)
	}

	return diskPath, vol, nil
}

func (a *LibvirtAdapter) parseAndValidatePort(p string) (int, int, error) {
	// Format: [hostPort:]containerPort
	parts := strings.Split(p, ":")
	var hostPort, containerPort string
	if len(parts) == 2 {
		hostPort = parts[0]
		containerPort = parts[1]
	} else if len(parts) == 1 {
		hostPort = "0"
		containerPort = parts[0]
	} else {
		return 0, 0, fmt.Errorf("invalid port format: too many colons")
	}

	// Security: Validate ports are numeric
	var hP, cP int
	_, errH := fmt.Sscanf(hostPort, "%d", &hP)
	_, errC := fmt.Sscanf(containerPort, "%d", &cP)
	if (hostPort != "0" && errH != nil) || errC != nil {
		return 0, 0, fmt.Errorf("invalid port format")
	}

	hPort := 0
	if hostPort == "0" {
		// Allocate random port (deterministic for simplicity in this POC)
		hPort = 30000 + int(uuid.New().ID()%10000)
	} else {
		// Parse host port, ignore error as we validate hPort > 0 below
		_, _ = fmt.Sscanf(hostPort, "%d", &hPort)
	}

	return hPort, cP, nil
}

func (a *LibvirtAdapter) resolveBinds(ctx context.Context, volumeBinds []string) []string {
	var additionalDisks []string
	if len(volumeBinds) == 0 {
		return additionalDisks
	}

	pool, poolErr := a.client.StoragePoolLookupByName(ctx, defaultPoolName)

	for _, bind := range volumeBinds {
		// Format is name:mountPath
		parts := strings.Split(bind, ":")
		volName := parts[0]

		if path := a.resolveVolumePath(ctx, volName, pool, poolErr); path != "" {
			additionalDisks = append(additionalDisks, path)
		}
	}
	return additionalDisks
}

func (a *LibvirtAdapter) resolveVolumePath(ctx context.Context, volName string, pool libvirt.StoragePool, poolErr error) string {
	// 1. Check if it's an LVM path
	if strings.HasPrefix(volName, "/dev/") {
		return volName
	}

	// 2. Try libvirt pool lookup (legacy/file-based)
	if poolErr == nil {
		v, err := a.client.StorageVolLookupByName(ctx, pool, volName)
		if err == nil {
			path, err := a.client.StorageVolGetPath(ctx, v)
			if err == nil {
				return path
			}
		}
	}

	// 3. Fallback: Check if it's a direct file path
	if strings.HasPrefix(volName, "/") {
		if _, statErr := os.Stat(volName); statErr == nil {
			return volName
		}
	}
	return ""
}
