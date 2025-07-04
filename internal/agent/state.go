package agent

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// GAState 全局状态管理 - 参考官方实现
type GAState struct {
	// 核心组件
	Parser       *JSONMessageParser
	MainLoop     *MainLoop
	Channel      Channel
	CommandState *CommandState

	// 状态标志
	LogLevel        logrus.Level
	LoggingEnabled  bool
	DelimitResponse bool
	Frozen          bool
	ForceExit       bool

	// 命令控制
	BlockedRPCs []string
	AllowedRPCs []string

	// 持久化状态
	PersistentState *PersistentState
	StateFilePath   string

	// 并发控制
	mutex sync.RWMutex
}

// PersistentState 持久化状态 - 参考官方实现
type PersistentState struct {
	FdCounter int64 `json:"fd_counter"`
}

// CommandState 命令状态管理
type CommandState struct {
	InitFunctions    []func()
	CleanupFunctions []func()
	mutex            sync.RWMutex
}

// JSONMessageParser JSON消息解析器
type JSONMessageParser struct {
	buffer     []byte
	tokenStart int
	inString   bool
	escaped    bool
	braceLevel int
}

// MainLoop 主循环管理
type MainLoop struct {
	running   bool
	stopChan  chan struct{}
	eventChan chan Event
	mutex     sync.RWMutex
}

// Event 事件类型
type Event struct {
	Type EventType
	Data interface{}
}

// EventType 事件类型枚举
type EventType int

const (
	EventMessage EventType = iota
	EventError
	EventShutdown
)

// NewGAState 创建新的全局状态
func NewGAState() *GAState {
	return &GAState{
		Parser:       NewJSONMessageParser(),
		MainLoop:     NewMainLoop(),
		CommandState: NewCommandState(),
		PersistentState: &PersistentState{
			FdCounter: 1000, // 默认文件描述符计数器起始值
		},
		LogLevel:       logrus.InfoLevel,
		LoggingEnabled: true,
	}
}

// NewJSONMessageParser 创建JSON消息解析器
func NewJSONMessageParser() *JSONMessageParser {
	return &JSONMessageParser{
		buffer: make([]byte, 0, 4096),
	}
}

// NewMainLoop 创建主循环
func NewMainLoop() *MainLoop {
	return &MainLoop{
		stopChan:  make(chan struct{}),
		eventChan: make(chan Event, 100),
	}
}

// NewCommandState 创建命令状态
func NewCommandState() *CommandState {
	return &CommandState{
		InitFunctions:    make([]func(), 0),
		CleanupFunctions: make([]func(), 0),
	}
}

// IsLoggingEnabled 检查是否启用日志
func (s *GAState) IsLoggingEnabled() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.LoggingEnabled
}

// SetLoggingEnabled 设置日志状态
func (s *GAState) SetLoggingEnabled(enabled bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.LoggingEnabled = enabled
}

// IsFrozen 检查是否处于冻结状态
func (s *GAState) IsFrozen() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.Frozen
}

// SetFrozen 设置冻结状态
func (s *GAState) SetFrozen(frozen bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Frozen = frozen
}

// SetResponseDelimited 设置响应分隔符
func (s *GAState) SetResponseDelimited(delimited bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.DelimitResponse = delimited
}

// IsResponseDelimited 检查是否使用响应分隔符
func (s *GAState) IsResponseDelimited() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.DelimitResponse
}

// GetFdHandle 获取文件描述符句柄 - 参考官方实现
func (s *GAState) GetFdHandle() int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.PersistentState.FdCounter++
	return s.PersistentState.FdCounter
}

// IsCommandAllowed 检查命令是否被允许
func (s *GAState) IsCommandAllowed(cmdName string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 如果处于冻结状态，只允许特定命令
	if s.Frozen {
		allowedWhenFrozen := []string{
			"guest-ping",
			"guest-info",
			"guest-sync",
			"guest-sync-delimited",
			"guest-fsfreeze-status",
			"guest-fsfreeze-thaw",
		}
		for _, allowed := range allowedWhenFrozen {
			if cmdName == allowed {
				return true
			}
		}
		return false
	}

	// 检查黑名单
	for _, blocked := range s.BlockedRPCs {
		if cmdName == blocked {
			return false
		}
	}

	// 如果有白名单，检查是否在白名单中
	if len(s.AllowedRPCs) > 0 {
		for _, allowed := range s.AllowedRPCs {
			if cmdName == allowed {
				return true
			}
		}
		return false
	}

	return true
}

// AddCommandStateInit 添加命令状态初始化函数
func (s *GAState) AddCommandStateInit(initFunc func()) {
	s.CommandState.mutex.Lock()
	defer s.CommandState.mutex.Unlock()
	s.CommandState.InitFunctions = append(s.CommandState.InitFunctions, initFunc)
}

// AddCommandStateCleanup 添加命令状态清理函数
func (s *GAState) AddCommandStateCleanup(cleanupFunc func()) {
	s.CommandState.mutex.Lock()
	defer s.CommandState.mutex.Unlock()
	s.CommandState.CleanupFunctions = append(s.CommandState.CleanupFunctions, cleanupFunc)
}

// InitCommandState 初始化命令状态
func (s *GAState) InitCommandState() {
	s.CommandState.mutex.RLock()
	defer s.CommandState.mutex.RUnlock()

	for _, initFunc := range s.CommandState.InitFunctions {
		initFunc()
	}
}

// CleanupCommandState 清理命令状态
func (s *GAState) CleanupCommandState() {
	s.CommandState.mutex.RLock()
	defer s.CommandState.mutex.RUnlock()

	for _, cleanupFunc := range s.CommandState.CleanupFunctions {
		cleanupFunc()
	}
}

// ParseMessage 解析JSON消息
func (p *JSONMessageParser) ParseMessage(data []byte) (*QMPMessage, error) {
	// 简单的JSON解析实现
	var message QMPMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return nil, err
	}
	return &message, nil
}

// QMPMessage QMP消息结构
type QMPMessage struct {
	Execute   string      `json:"execute,omitempty"`
	Arguments interface{} `json:"arguments,omitempty"`
	ID        interface{} `json:"id,omitempty"`
	Return    interface{} `json:"return,omitempty"`
	Error     *QMPError   `json:"error,omitempty"`
}

// QMPError QMP错误结构
type QMPError struct {
	Class string `json:"class"`
	Desc  string `json:"desc"`
}

// Start 启动主循环
func (m *MainLoop) Start() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.running {
		return
	}

	m.running = true
	go m.run()
}

// Stop 停止主循环
func (m *MainLoop) Stop() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.running {
		return
	}

	close(m.stopChan)
	m.running = false
}

// PostEvent 发送事件
func (m *MainLoop) PostEvent(event Event) {
	select {
	case m.eventChan <- event:
	default:
		logrus.Warn("事件队列已满，丢弃事件")
	}
}

// run 主循环执行
func (m *MainLoop) run() {
	for {
		select {
		case <-m.stopChan:
			return
		case event := <-m.eventChan:
			m.handleEvent(event)
		case <-time.After(time.Second):
			// 定期检查
			continue
		}
	}
}

// handleEvent 处理事件
func (m *MainLoop) handleEvent(event Event) {
	switch event.Type {
	case EventMessage:
		// 处理消息事件
		logrus.Debug("处理消息事件")
	case EventError:
		// 处理错误事件
		logrus.Error("处理错误事件")
	case EventShutdown:
		// 处理关闭事件
		logrus.Info("处理关闭事件")
		m.Stop()
	}
}
