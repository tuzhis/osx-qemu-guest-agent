package communication

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// TestManager 测试模式通信管理器，使用标准输入输出模拟virtio设备
type TestManager struct {
	reader   *bufio.Reader
	writer   *bufio.Writer
	isOpen   bool
	mutex    sync.RWMutex
	stopChan chan struct{}
}

// NewTestManager 创建测试模式通信管理器
func NewTestManager() *TestManager {
	return &TestManager{
		stopChan: make(chan struct{}),
	}
}

// Open 打开测试模式连接
func (m *TestManager) Open() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isOpen {
		return fmt.Errorf("测试设备已经打开")
	}

	m.reader = bufio.NewReader(os.Stdin)
	m.writer = bufio.NewWriter(os.Stdout)
	m.isOpen = true

	logrus.Info("测试模式: 使用标准输入输出模拟virtio设备")
	logrus.Info("测试模式: 你可以手动输入JSON命令进行测试")
	logrus.Info("测试模式: 示例命令:")
	logrus.Info(`  {"execute":"guest-ping"}`)
	logrus.Info(`  {"execute":"guest-info"}`)
	logrus.Info(`  {"execute":"guest-sync","arguments":{"id":12345}}`)
	
	return nil
}

// Close 关闭测试模式连接
func (m *TestManager) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isOpen {
		return nil
	}

	close(m.stopChan)
	m.isOpen = false
	logrus.Info("测试模式连接已关闭")
	return nil
}

// ReadMessage 从标准输入读取消息
func (m *TestManager) ReadMessage() ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.isOpen {
		return nil, fmt.Errorf("测试设备未打开")
	}

	fmt.Print("请输入QMP命令 > ")
	
	// 读取一行数据
	line, err := m.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("读取输入失败: %v", err)
	}

	// 清理换行符
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("输入为空")
	}

	// 支持退出命令
	if line == "quit" || line == "exit" {
		return nil, fmt.Errorf("用户退出")
	}

	logrus.WithField("input", line).Debug("收到测试输入")
	return []byte(line), nil
}

// SendResponse 向标准输出发送响应
func (m *TestManager) SendResponse(data []byte) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.isOpen {
		return fmt.Errorf("测试设备未打开")
	}

	// 格式化输出
	fmt.Printf("QMP响应: %s\n", string(data))
	
	err := m.writer.Flush()
	if err != nil {
		return fmt.Errorf("刷新输出缓冲区失败: %v", err)
	}

	return nil
}

// SendDelimitedResponse 向标准输出发送带分隔符的响应（测试模式）
func (m *TestManager) SendDelimitedResponse(data []byte) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.isOpen {
		return fmt.Errorf("测试设备未打开")
	}

	// 格式化输出（测试模式显示分隔符）
	fmt.Printf("QMP带分隔符响应[0xFF]: %s\n", string(data))
	
	err := m.writer.Flush()
	if err != nil {
		return fmt.Errorf("刷新输出缓冲区失败: %v", err)
	}

	return nil
}

// IsOpen 检查测试设备是否已打开
func (m *TestManager) IsOpen() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.isOpen
} 