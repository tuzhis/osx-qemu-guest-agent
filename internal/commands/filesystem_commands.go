package commands

import (
	"bufio"
	"encoding/json"
	"mac-guest-agent/internal/protocol"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	// freezeStatus simulates the filesystem freeze status, as macOS does not
	// have a native equivalent to Linux's `fsfreeze`.
	freezeStatus = protocol.FsfreezeStatusThawed
)

func init() {
	RegisterCommand(&Command{Name: "guest-get-fsinfo", Handler: handleGetFSInfo, Enabled: true})
	RegisterCommand(&Command{Name: "guest-fsfreeze-status", Handler: handleFSFreezeStatus, Enabled: true})
	RegisterCommand(&Command{Name: "guest-fsfreeze-freeze", Handler: handleFSFreezeFreeze, Enabled: true})
	RegisterCommand(&Command{Name: "guest-fsfreeze-thaw", Handler: handleFSFreezeThaw, Enabled: true})
	RegisterCommand(&Command{Name: "guest-fstrim", Handler: handleFSTrim, Enabled: true})
}

// handleGetFSInfo handles the guest-get-fsinfo command.
func handleGetFSInfo(req json.RawMessage) (interface{}, error) {
	filesystems, err := getFilesystemInfo()
	if err != nil {
		logrus.WithError(err).Error("Failed to get filesystem info")
		return nil, err
	}
	logrus.WithField("filesystem_count", len(filesystems)).Info("Successfully retrieved filesystem info")
	return filesystems, nil
}

// handleFSFreezeStatus handles the guest-fsfreeze-status command.
func handleFSFreezeStatus(req json.RawMessage) (interface{}, error) {
	logrus.WithField("status", freezeStatus).Info("Returning simulated filesystem freeze status")
	return freezeStatus, nil
}

// handleFSFreezeFreeze handles the guest-fsfreeze-freeze command.
func handleFSFreezeFreeze(req json.RawMessage) (interface{}, error) {
	logrus.Info("Simulating filesystem freeze. No actual freeze is performed on macOS.")
	freezeStatus = protocol.FsfreezeStatusFrozen
	// Return the number of "frozen" filesystems, which is always 1 for the root.
	return 1, nil
}

// handleFSFreezeThaw handles the guest-fsfreeze-thaw command.
func handleFSFreezeThaw(req json.RawMessage) (interface{}, error) {
	thawedCount := 0
	if freezeStatus == protocol.FsfreezeStatusFrozen {
		thawedCount = 1
		freezeStatus = protocol.FsfreezeStatusThawed
		logrus.Info("Simulating filesystem thaw.")
	}
	return thawedCount, nil
}

// handleFSTrim handles the guest-fstrim command.
func handleFSTrim(req json.RawMessage) (interface{}, error) {
	logrus.Info("guest-fstrim is a no-op on macOS as TRIM is managed by the OS and storage driver.")
	return protocol.GuestFilesystemTrimResponse{Paths: []protocol.GuestFilesystemTrimResult{}}, nil
}

// getFilesystemInfo retrieves information about mounted filesystems.
func getFilesystemInfo() ([]protocol.GuestFilesystemInfo, error) {
	cmd := exec.Command("df", "-kP") // Use -k for consistent kilobyte blocks, -P for POSIX format
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseDfOutput(string(output))
}

// parseDfOutput parses the output of the `df -kP` command.
func parseDfOutput(output string) ([]protocol.GuestFilesystemInfo, error) {
	var filesystems []protocol.GuestFilesystemInfo
	scanner := bufio.NewScanner(strings.NewReader(output))
	scanner.Scan() // Skip header line

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		deviceName := fields[0]
		if !strings.HasPrefix(deviceName, "/dev/") {
			continue // Skip virtual/network filesystems
		}

		total, _ := strconv.ParseInt(fields[1], 10, 64)
		used, _ := strconv.ParseInt(fields[2], 10, 64)

		fs := protocol.GuestFilesystemInfo{
			Name:       deviceName,
			Mountpoint: fields[5],
			Type:       getFilesystemType(deviceName),
			TotalBytes: total * 1024,
			UsedBytes:  used * 1024,
		}
		filesystems = append(filesystems, fs)
	}
	return filesystems, nil
}

// getFilesystemType retrieves the filesystem type using the `mount` command.
func getFilesystemType(deviceName string) string {
	cmd := exec.Command("mount")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	re := regexp.MustCompile(`^` + regexp.QuoteMeta(deviceName) + ` on .* \(([^,]+)`)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		matches := re.FindStringSubmatch(scanner.Text())
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return "unknown"
}
