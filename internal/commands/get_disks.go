package commands

import (
	"encoding/json"
	"mac-guest-agent/internal/protocol"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterCommand(&Command{
		Name:    "guest-get-disks",
		Handler: handleGetDisks,
		Enabled: true,
	})
}

// handleGetDisks handles the guest-get-disks command.
func handleGetDisks(req json.RawMessage) (interface{}, error) {
	disks, err := getDisks()
	if err != nil {
		logrus.WithError(err).Error("Failed to get disk information")
		return nil, err
	}
	logrus.WithField("disk_count", len(disks)).Info("Successfully retrieved disk information")
	return disks, nil
}

// getDisks retrieves information about disks.
func getDisks() ([]protocol.GuestDiskInfo, error) {
	// On macOS, use diskutil list to get disk information.
	cmd := exec.Command("diskutil", "list", "-plist")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Since parsing plist is complex without additional dependencies,
	// we'll use diskutil list in text format instead.
	cmd = exec.Command("diskutil", "list")
	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseDiskUtilOutput(string(output))
}

// parseDiskUtilOutput parses the output of diskutil list.
func parseDiskUtilOutput(output string) ([]protocol.GuestDiskInfo, error) {
	var disks []protocol.GuestDiskInfo
	var currentDisk *protocol.GuestDiskInfo

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this is a new disk
		if strings.HasPrefix(line, "/dev/disk") && !strings.Contains(line, "s") {
			// If we were processing a disk, add it to the list
			if currentDisk != nil {
				disks = append(disks, *currentDisk)
			}

			// Start a new disk
			diskName := strings.Split(line, " ")[0]
			currentDisk = &protocol.GuestDiskInfo{
				Name:      diskName,
				Partition: false,
				HasMedia:  true,
				Address: &protocol.GuestDiskAddress{
					BusType: protocol.DiskBusUnknown,
					Bus:     -1,
					Target:  -1,
					Unit:    -1,
					PCIController: protocol.GuestPCIAddress{
						Domain:   -1,
						Bus:      -1,
						Slot:     -1,
						Function: -1,
					},
					Dev: diskName,
				},
			}

			// Try to get disk size
			if sizeStr := getDiskSize(diskName); sizeStr != "" {
				if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
					currentDisk.Size = size
				}
			}
		} else if currentDisk != nil && strings.Contains(line, ":") {
			// This might be a partition
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 && strings.Contains(parts[0], " ") {
				numParts := strings.Fields(parts[0])
				if len(numParts) > 0 {
					numStr := strings.Trim(numParts[0], " ")
					if num, err := strconv.Atoi(numStr); err == nil {
						name := strings.TrimSpace(parts[1])
						partition := protocol.GuestPartitionInfo{
							Number: num,
							Name:   name,
						}
						currentDisk.Partitions = append(currentDisk.Partitions, partition)
					}
				}
			}
		}
	}

	// Add the last disk if we were processing one
	if currentDisk != nil {
		disks = append(disks, *currentDisk)
	}

	return disks, nil
}

// getDiskSize retrieves the size of a disk in bytes.
func getDiskSize(diskName string) string {
	cmd := exec.Command("diskutil", "info", "-plist", diskName)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Since parsing plist is complex without additional dependencies,
	// we'll use diskutil info in text format instead.
	cmd = exec.Command("diskutil", "info", diskName)
	output, err = cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Disk Size") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				sizeStr := strings.TrimSpace(parts[1])
				// Extract the bytes part from something like "500.3 GB (500277108736 Bytes)"
				if strings.Contains(sizeStr, "(") && strings.Contains(sizeStr, "Bytes)") {
					bytesStr := strings.Split(strings.Split(sizeStr, "(")[1], " ")[0]
					return bytesStr
				}
			}
		}
	}

	return ""
}
