package commands

import (
	"bufio"
	"bytes"
	"encoding/json"
	"mac-guest-agent/internal/protocol"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

var (
	swVersOnce   sync.Once
	swVersOutput map[string]string
)

func init() {
	RegisterCommand(&Command{
		Name:    "guest-get-osinfo",
		Handler: handleGetOSInfo,
		Enabled: true,
	})
}

// handleGetOSInfo handles the guest-get-osinfo command.
func handleGetOSInfo(req json.RawMessage) (interface{}, error) {
	osInfo := &protocol.GuestOSInfo{
		ID:         "macos",
		Name:       "macOS",
		Variant:    "desktop",
		VariantID:  "desktop",
		PrettyName: getOSInfoField("ProductName") + " " + getOSInfoField("ProductVersion"),
		Version:    getOSInfoField("ProductVersion"),
		VersionID:  getOSInfoField("BuildVersion"),
	}

	if uname, err := getUnameInfo(); err == nil {
		osInfo.KernelRelease = uname.Release
		osInfo.KernelVersion = uname.Version
		osInfo.Machine = uname.Machine
	}

	logrus.WithFields(logrus.Fields{
		"id":          osInfo.ID,
		"pretty_name": osInfo.PrettyName,
		"version":     osInfo.Version,
		"kernel":      osInfo.KernelRelease,
	}).Info("Successfully retrieved OS information")

	return osInfo, nil
}

// UnameInfo holds system information from uname.
type UnameInfo struct {
	Release string
	Version string
	Machine string
}

// getUnameInfo retrieves uname information using syscalls.
func getUnameInfo() (*UnameInfo, error) {
	var utsname unix.Utsname
	if err := unix.Uname(&utsname); err != nil {
		return nil, err
	}
	// Helper to convert C-style byte arrays to Go strings.
	cToGo := func(c []byte) string {
		n := bytes.IndexByte(c, 0)
		if n < 0 {
			n = len(c)
		}
		return string(c[:n])
	}
	info := &UnameInfo{
		Release: cToGo(utsname.Release[:]),
		Version: cToGo(utsname.Version[:]),
		Machine: cToGo(utsname.Machine[:]),
	}
	return info, nil
}

// getOSInfoField retrieves a specific field from the `sw_vers` command output.
// It calls the command only once and caches the result for subsequent calls.
func getOSInfoField(field string) string {
	swVersOnce.Do(func() {
		swVersOutput = make(map[string]string)
		if runtime.GOOS != "darwin" {
			return
		}
		cmd := exec.Command("sw_vers")
		output, err := cmd.Output()
		if err != nil {
			return
		}
		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				swVersOutput[key] = value
			}
		}
	})
	return swVersOutput[field]
}
