package communication

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Manager 通信管理器
type Manager struct {
	devicePath string
	device     *os.File
	reader     *bufio.Reader
	writer     *bufio.Writer
	isOpen     bool
	mutex      sync.RWMutex
	stopChan   chan struct{}
}

// NewManager 创建新的通信管理器
func NewManager(devicePath string) *Manager {
	return &Manager{
		devicePath: devicePath,
		stopChan:   make(chan struct{}),
	}
}

// DetectDevice 自动检测virtio设备
func DetectDevice() (string, error) {
	possiblePaths := []string{
		// QEMU Guest Agent 标准设备路径
		"/dev/cu.org.qemu.guest_agent.0",
		"/dev/tty.org.qemu.guest_agent.0",
		// 通用virtio设备路径
		"/dev/cu.virtio-console.0",
		"/dev/tty.virtio-console.0", 
		"/dev/cu.virtio-serial",
		"/dev/tty.virtio-serial",
		"/dev/cu.virtio-port",
		"/dev/tty.virtio-port",
		// 其他可能的QEMU设备路径
		"/dev/cu.qemu-guest-agent",
		"/dev/tty.qemu-guest-agent",
	}

	logrus.Debug("正在检测virtio设备...")
	for _, path := range possiblePaths {
		logrus.WithField("path", path).Debug("检查设备路径")
		if stat, err := os.Stat(path); err == nil {
			// 检查是否为字符设备
			if stat.Mode()&os.ModeCharDevice != 0 {
				logrus.WithField("device", path).Info("检测到virtio设备")
				return path, nil
			} else {
				logrus.WithField("path", path).Debug("路径存在但不是字符设备")
			}
		} else {
			logrus.WithField("path", path).WithError(err).Debug("设备路径不存在")
		}
	}

	return "", fmt.Errorf("未找到可用的virtio设备，已检查的路径: %v", possiblePaths)
}

// Open 打开设备连接
func (m *Manager) Open() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isOpen {
		return fmt.Errorf("设备已经打开")
	}

	// 如果没有指定设备路径，自动检测
	if m.devicePath == "" {
		path, err := DetectDevice()
		if err != nil {
			return fmt.Errorf("检测设备失败: %v", err)
		}
		m.devicePath = path
	}

	// 打开设备文件
	device, err := os.OpenFile(m.devicePath, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("打开设备 %s 失败: %v", m.devicePath, err)
	}

	m.device = device
	m.reader = bufio.NewReader(device)
	m.writer = bufio.NewWriter(device)
	m.isOpen = true

	logrus.WithField("device", m.devicePath).Info("成功打开virtio设备")
	return nil
}

// Close 关闭设备连接
func (m *Manager) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isOpen {
		return nil
	}

	close(m.stopChan)

	if m.device != nil {
		m.device.Close()
	}

	m.isOpen = false
	logrus.Info("已关闭virtio设备连接")
	return nil
}

// ReadMessage 读取消息
func (m *Manager) ReadMessage() ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.isOpen {
		return nil, fmt.Errorf("设备未打开")
	}

	// 设置读取超时 - 使用短超时以便快速响应退出信号
	m.device.SetReadDeadline(time.Now().Add(1 * time.Second))

	// 读取一行数据（JSON消息以换行符结束）
	line, err := m.reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("设备连接已关闭")
		}
		// 检查是否是超时错误
		if strings.Contains(err.Error(), "timeout") {
			// Guest Agent的正常工作模式就是等待命令，超时是正常的
			return nil, fmt.Errorf("read_timeout")
		}
		return nil, fmt.Errorf("读取消息失败: %v", err)
	}

	// 清理换行符和空白字符
	line = strings.TrimSpace(line)
	if line == "" {
		// 收到空行，继续等待下一条消息
		return nil, fmt.Errorf("empty_message")
	}

	logrus.WithField("message", line).Debug("收到消息")
	return []byte(line), nil
}

// SendResponse 发送响应
func (m *Manager) SendResponse(data []byte) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.isOpen {
		return fmt.Errorf("设备未打开")
	}

	// 设置写入超时
	m.device.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// 写入数据（添加换行符）
	_, err := m.writer.Write(data)
	if err != nil {
		return fmt.Errorf("写入响应失败: %v", err)
	}

	_, err = m.writer.WriteString("\n")
	if err != nil {
		return fmt.Errorf("写入换行符失败: %v", err)
	}

	// 刷新缓冲区
	err = m.writer.Flush()
	if err != nil {
		return fmt.Errorf("刷新缓冲区失败: %v", err)
	}

	logrus.WithField("response", string(data)).Debug("发送响应")
	return nil
}

// SendDelimitedResponse 发送带分隔符的响应（用于guest-sync-delimited）
func (m *Manager) SendDelimitedResponse(data []byte) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.isOpen {
		return fmt.Errorf("设备未打开")
	}

	// 设置写入超时
	m.device.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// 先发送0xFF分隔符
	_, err := m.writer.Write([]byte{0xFF})
	if err != nil {
		return fmt.Errorf("写入分隔符失败: %v", err)
	}

	// 写入数据（添加换行符）
	_, err = m.writer.Write(data)
	if err != nil {
		return fmt.Errorf("写入响应失败: %v", err)
	}

	_, err = m.writer.WriteString("\n")
	if err != nil {
		return fmt.Errorf("写入换行符失败: %v", err)
	}

	// 刷新缓冲区
	err = m.writer.Flush()
	if err != nil {
		return fmt.Errorf("刷新缓冲区失败: %v", err)
	}

	logrus.WithField("response", string(data)).Debug("发送带分隔符的响应")
	return nil
}

// IsOpen 检查设备是否已打开
func (m *Manager) IsOpen() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.isOpen
} 