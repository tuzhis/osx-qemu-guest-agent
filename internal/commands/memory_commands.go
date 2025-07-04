package commands

import (
	"bufio"
	"encoding/json"
	"mac-guest-agent/internal/protocol"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterCommand(&Command{
		Name:    "guest-get-memory-blocks",
		Handler: handleGetMemoryBlocks,
		Enabled: true,
	})
	RegisterCommand(&Command{
		Name:    "guest-get-memory-block-info",
		Handler: handleGetMemoryBlockInfo,
		Enabled: true,
	})
	// guest-get-memory-info is an alias in some agents
	RegisterCommand(&Command{
		Name:    "guest-get-memory-info",
		Handler: handleGetMemoryInfo,
		Enabled: true,
	})
	RegisterCommand(&Command{
		Name:    "guest-set-memory-blocks",
		Handler: handleSetMemoryBlocks,
		Enabled: true,
	})
}

// handleGetMemoryBlocks handles the guest-get-memory-blocks command.
func handleGetMemoryBlocks(req json.RawMessage) (interface{}, error) {
	blocks, err := getMemoryBlocks()
	if err != nil {
		logrus.WithError(err).Error("Failed to get memory blocks")
		return nil, err
	}
	logrus.WithField("block_count", len(blocks)).Info("Successfully retrieved memory blocks")
	return blocks, nil
}

// handleGetMemoryBlockInfo handles the guest-get-memory-block-info command.
func handleGetMemoryBlockInfo(req json.RawMessage) (interface{}, error) {
	info, err := getMemoryBlockInfo()
	if err != nil {
		logrus.WithError(err).Error("Failed to get memory block info")
		return nil, err
	}
	logrus.WithField("block_size", info.Size).Info("Successfully retrieved memory block info")
	return info, nil
}

// handleGetMemoryInfo handles the guest-get-memory-info command.
// This is a more comprehensive command that is not part of the standard,
// but useful for macOS.
func handleGetMemoryInfo(req json.RawMessage) (interface{}, error) {
	return getMemoryInfo()
}

// handleSetMemoryBlocks handles the guest-set-memory-blocks command.
// This is a no-op on macOS as memory hotplug is not supported.
func handleSetMemoryBlocks(req json.RawMessage) (interface{}, error) {
	logrus.Warn("guest-set-memory-blocks is not supported on macOS")
	return protocol.EmptyResponse{}, nil
}

// getMemoryBlocks retrieves information about memory blocks.
func getMemoryBlocks() ([]protocol.GuestMemoryBlock, error) {
	// macOS does not support memory hot-plugging in the same way as Linux.
	// This provides a simulated implementation.
	totalMemory, err := getTotalMemory()
	if err != nil {
		return nil, err
	}

	// 动态计算内存块大小，而不是固定为1GB
	// 对于小内存系统（<4GB），使用256MB块
	// 对于中等内存系统（4-16GB），使用512MB块
	// 对于大内存系统（>16GB），使用1GB块
	// 这样可以提供更合理的内存块数量
	var blockSize int64
	const (
		MB = 1024 * 1024
		GB = 1024 * MB
	)

	switch {
	case totalMemory < 4*GB:
		blockSize = 256 * MB
	case totalMemory < 16*GB:
		blockSize = 512 * MB
	default:
		blockSize = 1 * GB
	}

	// 确保块数量在合理范围内（8-32块）
	numBlocks := int(totalMemory / blockSize)
	if totalMemory%blockSize > 0 {
		numBlocks++
	}

	// 调整块大小以保持合理的块数量
	if numBlocks < 8 {
		// 如果块数太少，减小块大小
		blockSize = totalMemory / 8
		if blockSize < 128*MB {
			blockSize = 128 * MB
		}
		numBlocks = int(totalMemory / blockSize)
		if totalMemory%blockSize > 0 {
			numBlocks++
		}
	} else if numBlocks > 32 {
		// 如果块数太多，增加块大小
		blockSize = totalMemory / 32
		numBlocks = int(totalMemory / blockSize)
		if totalMemory%blockSize > 0 {
			numBlocks++
		}
	}

	if numBlocks == 0 {
		numBlocks = 1
	}

	logrus.WithFields(logrus.Fields{
		"total_memory": totalMemory,
		"block_size":   blockSize,
		"num_blocks":   numBlocks,
	}).Debug("Calculated memory blocks")

	var blocks []protocol.GuestMemoryBlock
	for i := 0; i < numBlocks; i++ {
		block := protocol.GuestMemoryBlock{
			PhysIndex:  i,
			Online:     true,
			CanOffline: false, // Memory hot-unplug is not supported on macOS.
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

// getMemoryBlockInfo retrieves information about memory block size.
func getMemoryBlockInfo() (*protocol.GuestMemoryBlockInfo, error) {
	// 获取动态计算的内存块大小，与getMemoryBlocks保持一致
	totalMemory, err := getTotalMemory()
	if err != nil {
		return nil, err
	}

	// 使用与getMemoryBlocks相同的逻辑计算块大小
	var blockSize int64
	const (
		MB = 1024 * 1024
		GB = 1024 * MB
	)

	switch {
	case totalMemory < 4*GB:
		blockSize = 256 * MB
	case totalMemory < 16*GB:
		blockSize = 512 * MB
	default:
		blockSize = 1 * GB
	}

	// 调整块大小以保持合理的块数量
	numBlocks := int(totalMemory / blockSize)
	if totalMemory%blockSize > 0 {
		numBlocks++
	}

	if numBlocks < 8 {
		blockSize = totalMemory / 8
		if blockSize < 128*MB {
			blockSize = 128 * MB
		}
	} else if numBlocks > 32 {
		blockSize = totalMemory / 32
	}

	info := &protocol.GuestMemoryBlockInfo{
		Size: blockSize,
	}
	return info, nil
}

// getTotalMemory retrieves the total system memory.
func getTotalMemory() (int64, error) {
	// On macOS, use sysctl to get memory information.
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	memSizeStr := strings.TrimSpace(string(output))
	return strconv.ParseInt(memSizeStr, 10, 64)
}

// getMemoryInfo retrieves detailed memory information from `vm_stat`.
func getMemoryInfo() (map[string]int64, error) {
	cmd := exec.Command("vm_stat")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		valueStr := strings.TrimRight(strings.TrimSpace(parts[1]), ".")
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			continue
		}
		stats[key] = value
	}
	return stats, nil
}
