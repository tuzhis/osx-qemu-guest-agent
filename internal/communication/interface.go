package communication

// CommunicationManager 通信管理器接口
type CommunicationManager interface {
	Open() error
	Close() error
	ReadMessage() ([]byte, error)
	SendResponse(data []byte) error
	SendDelimitedResponse(data []byte) error
	IsOpen() bool
} 