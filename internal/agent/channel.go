package agent

import (
	"io"
)

// Channel 通信通道接口 - 参考官方实现
type Channel interface {
	// Open 打开通道
	Open() error

	// Close 关闭通道
	Close() error

	// Read 读取数据
	Read(buffer []byte) (int, error)

	// Write 写入数据
	Write(data []byte) (int, error)

	// WriteAll 写入所有数据
	WriteAll(data []byte) error

	// IsOpen 检查通道是否打开
	IsOpen() bool

	// GetPath 获取通道路径
	GetPath() string
}

// ChannelMethod 通道方法类型
type ChannelMethod int

const (
	ChannelVirtioSerial ChannelMethod = iota
	ChannelISASerial
	ChannelUnixListen
	ChannelVsockListen
	ChannelTest
)

// ChannelCallback 通道回调函数类型
type ChannelCallback func(condition IOCondition, data interface{}) bool

// IOCondition IO条件标志
type IOCondition int

const (
	IOConditionIn IOCondition = 1 << iota
	IOConditionOut
	IOConditionPri
	IOConditionErr
	IOConditionHup
	IOConditionNval
)

// ChannelConfig 通道配置
type ChannelConfig struct {
	Method   ChannelMethod
	Path     string
	ListenFd int
	Callback ChannelCallback
	Opaque   interface{}
}

// BaseChannel 基础通道实现
type BaseChannel struct {
	config *ChannelConfig
	isOpen bool
	reader io.Reader
	writer io.Writer
	closer io.Closer
}

// NewChannel 创建新的通道
func NewChannel(config *ChannelConfig) (Channel, error) {
	switch config.Method {
	case ChannelVirtioSerial:
		return NewVirtioSerialChannel(config)
	case ChannelISASerial:
		return NewISASerialChannel(config)
	case ChannelUnixListen:
		return NewUnixListenChannel(config)
	case ChannelVsockListen:
		return NewVsockListenChannel(config)
	case ChannelTest:
		return NewTestChannel(config)
	default:
		return nil, ErrUnsupportedChannelMethod
	}
}

// ErrUnsupportedChannelMethod 不支持的通道方法错误
var ErrUnsupportedChannelMethod = &ChannelError{
	Code:    "UnsupportedChannelMethod",
	Message: "不支持的通道方法",
}

// ChannelError 通道错误
type ChannelError struct {
	Code    string
	Message string
}

func (e *ChannelError) Error() string {
	return e.Code + ": " + e.Message
}

// Open 打开通道
func (c *BaseChannel) Open() error {
	if c.isOpen {
		return &ChannelError{
			Code:    "ChannelAlreadyOpen",
			Message: "通道已经打开",
		}
	}
	c.isOpen = true
	return nil
}

// Close 关闭通道
func (c *BaseChannel) Close() error {
	if !c.isOpen {
		return nil
	}

	if c.closer != nil {
		if err := c.closer.Close(); err != nil {
			return err
		}
	}

	c.isOpen = false
	return nil
}

// IsOpen 检查通道是否打开
func (c *BaseChannel) IsOpen() bool {
	return c.isOpen
}

// GetPath 获取通道路径
func (c *BaseChannel) GetPath() string {
	if c.config == nil {
		return ""
	}
	return c.config.Path
}

// Read 读取数据
func (c *BaseChannel) Read(buffer []byte) (int, error) {
	if !c.isOpen {
		return 0, &ChannelError{
			Code:    "ChannelNotOpen",
			Message: "通道未打开",
		}
	}

	if c.reader == nil {
		return 0, &ChannelError{
			Code:    "NoReader",
			Message: "没有可用的读取器",
		}
	}

	return c.reader.Read(buffer)
}

// Write 写入数据
func (c *BaseChannel) Write(data []byte) (int, error) {
	if !c.isOpen {
		return 0, &ChannelError{
			Code:    "ChannelNotOpen",
			Message: "通道未打开",
		}
	}

	if c.writer == nil {
		return 0, &ChannelError{
			Code:    "NoWriter",
			Message: "没有可用的写入器",
		}
	}

	return c.writer.Write(data)
}

// WriteAll 写入所有数据
func (c *BaseChannel) WriteAll(data []byte) error {
	if !c.isOpen {
		return &ChannelError{
			Code:    "ChannelNotOpen",
			Message: "通道未打开",
		}
	}

	written := 0
	for written < len(data) {
		n, err := c.Write(data[written:])
		if err != nil {
			return err
		}
		written += n
	}

	return nil
}

// VirtioSerialChannel Virtio串口通道
type VirtioSerialChannel struct {
	*BaseChannel
}

// NewVirtioSerialChannel 创建Virtio串口通道
func NewVirtioSerialChannel(config *ChannelConfig) (Channel, error) {
	return &VirtioSerialChannel{
		BaseChannel: &BaseChannel{
			config: config,
		},
	}, nil
}

// ISASerialChannel ISA串口通道
type ISASerialChannel struct {
	*BaseChannel
}

// NewISASerialChannel 创建ISA串口通道
func NewISASerialChannel(config *ChannelConfig) (Channel, error) {
	return &ISASerialChannel{
		BaseChannel: &BaseChannel{
			config: config,
		},
	}, nil
}

// UnixListenChannel Unix监听通道
type UnixListenChannel struct {
	*BaseChannel
}

// NewUnixListenChannel 创建Unix监听通道
func NewUnixListenChannel(config *ChannelConfig) (Channel, error) {
	return &UnixListenChannel{
		BaseChannel: &BaseChannel{
			config: config,
		},
	}, nil
}

// VsockListenChannel Vsock监听通道
type VsockListenChannel struct {
	*BaseChannel
}

// NewVsockListenChannel 创建Vsock监听通道
func NewVsockListenChannel(config *ChannelConfig) (Channel, error) {
	return &VsockListenChannel{
		BaseChannel: &BaseChannel{
			config: config,
		},
	}, nil
}

// TestChannel 测试通道
type TestChannel struct {
	*BaseChannel
}

// NewTestChannel 创建测试通道
func NewTestChannel(config *ChannelConfig) (Channel, error) {
	return &TestChannel{
		BaseChannel: &BaseChannel{
			config: config,
		},
	}, nil
}
