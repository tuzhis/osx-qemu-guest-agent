package commands

import (
	"encoding/json"
	"mac-guest-agent/internal/protocol"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterCommand(&Command{
		Name:    "guest-get-vcpus",
		Handler: handleGetVCPUs,
		Enabled: true,
	})
}

// handleGetVCPUs handles the guest-get-vcpus command.
func handleGetVCPUs(req json.RawMessage) (interface{}, error) {
	vcpus, err := getVirtualCPUs()
	if err != nil {
		logrus.WithError(err).Error("Failed to get vCPU information")
		return nil, err
	}
	logrus.WithField("vcpu_count", len(vcpus)).Info("Successfully retrieved vCPU information")
	return vcpus, nil
}

// getVirtualCPUs retrieves information about the virtual CPUs.
// On macOS, all logical processors are reported as online and cannot be
// hot-unplugged.
func getVirtualCPUs() ([]protocol.GuestLogicalProcessor, error) {
	numCPU := runtime.NumCPU()
	vcpus := make([]protocol.GuestLogicalProcessor, numCPU)

	for i := 0; i < numCPU; i++ {
		vcpus[i] = protocol.GuestLogicalProcessor{
			LogicalID:  i,
			Online:     true,
			CanOffline: false,
		}
	}

	// 尝试从系统获取更详细的CPU信息
	if detailedVCPUs := getDetailedCPUInfo(); detailedVCPUs != nil {
		return detailedVCPUs, nil
	}

	return vcpus, nil
}

// getDetailedCPUInfo 获取详细的CPU信息
func getDetailedCPUInfo() []protocol.GuestLogicalProcessor {
	// 在macOS上使用sysctl获取CPU信息
	cmd := exec.Command("sysctl", "-n", "machdep.cpu.thread_count")
	output, err := cmd.Output()
	if err != nil {
		logrus.WithError(err).Debug("无法获取CPU线程数")
		return nil
	}

	threadCountStr := strings.TrimSpace(string(output))
	threadCount, err := strconv.Atoi(threadCountStr)
	if err != nil {
		logrus.WithError(err).Debug("解析CPU线程数失败")
		return nil
	}

	var vcpus []protocol.GuestLogicalProcessor
	for i := 0; i < threadCount; i++ {
		vcpu := protocol.GuestLogicalProcessor{
			LogicalID:  i,
			Online:     true,
			CanOffline: false,
		}
		vcpus = append(vcpus, vcpu)
	}

	return vcpus
}
