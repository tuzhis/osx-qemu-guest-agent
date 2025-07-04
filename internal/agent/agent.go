package agent

import (
	"encoding/json"
	"fmt"
	"mac-guest-agent/internal/commands"
	"mac-guest-agent/internal/communication"
	"mac-guest-agent/internal/protocol"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	Version = "1.1.0"
)

// Agent represents the main class for the macOS Guest Agent.
type Agent struct {
	commManager communication.CommunicationManager
	isRunning   bool
	stopChan    chan struct{}
	mutex       sync.RWMutex
}

// New creates a new Agent instance.
func New(devicePath string) (*Agent, error) {
	agent := &Agent{
		commManager: communication.NewManager(devicePath),
		stopChan:    make(chan struct{}),
	}
	return agent, nil
}

// NewTestMode creates a test mode Agent instance.
func NewTestMode() (*Agent, error) {
	agent := &Agent{
		commManager: communication.NewTestManager(),
		stopChan:    make(chan struct{}),
	}
	return agent, nil
}

// Start starts the Agent.
func (a *Agent) Start() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.isRunning {
		return fmt.Errorf("Agent is already running")
	}

	if err := a.commManager.Open(); err != nil {
		return fmt.Errorf("failed to open communication device: %v", err)
	}

	a.isRunning = true
	logrus.Info("Agent started, listening for messages...")

	go a.messageLoop()

	return nil
}

// Stop stops the Agent.
func (a *Agent) Stop() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.isRunning {
		return
	}

	close(a.stopChan)
	a.commManager.Close()
	a.isRunning = false

	logrus.Info("Agent stopped")
}

// IsRunning checks if the Agent is running.
func (a *Agent) IsRunning() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.isRunning
}

// messageLoop is the main message processing loop.
func (a *Agent) messageLoop() {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithField("panic", r).Error("Message processing loop panicked")
		}
	}()

	for {
		select {
		case <-a.stopChan:
			logrus.Debug("Received stop signal, exiting message loop")
			return
		default:
			if err := a.processMessage(); err != nil {
				if err.Error() == "read_timeout" || err.Error() == "empty_message" || strings.Contains(err.Error(), "timeout") {
					continue
				}

				logrus.WithError(err).Error("Failed to process message")

				if !a.commManager.IsOpen() {
					logrus.Info("Device connection lost, attempting to reconnect...")
					time.Sleep(5 * time.Second)
					if err := a.commManager.Open(); err != nil {
						logrus.WithError(err).Error("Failed to reconnect")
					}
				} else {
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}
}

// processMessage handles a single message.
func (a *Agent) processMessage() error {
	msgData, err := a.commManager.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read message: %v", err)
	}

	request, err := protocol.ParseRequest(msgData)
	if err != nil {
		logrus.WithError(err).WithField("data", string(msgData)).Error("Failed to parse request")
		errorResp := protocol.NewErrorResponse("GenericError", "Invalid message format")
		return a.sendResponse(errorResp, false)
	}

	// Log arguments safely
	var argsStr string
	if request.Arguments != nil {
		argBytes, marshalErr := json.Marshal(request.Arguments)
		if marshalErr != nil {
			argsStr = "[failed to marshal arguments]"
		} else {
			argsStr = string(argBytes)
		}
	}

	logFields := logrus.Fields{
		"command":   request.Execute,
		"arguments": argsStr,
		"id":        request.ID,
	}

	// Reduce log noise for frequent commands.
	if request.Execute == "guest-ping" || request.Execute == "guest-sync-delimited" {
		logrus.WithFields(logFields).Debug("Received QMP request")
	} else {
		logrus.WithFields(logFields).Info("Received QMP request")
	}

	// The old processor is gone, directly call the new command handler.
	// The new handler returns a `protocol.Response` which needs to be converted.
	handlerResponse := commands.HandleCommand(*request)
	qmpResponse := protocol.QMPResponse{
		Return: handlerResponse.Return,
		Error:  handlerResponse.Error,
	}

	// The response ID must match the request ID.
	if request.ID != nil {
		qmpResponse.ID = request.ID
	}

	// For guest-sync-delimited, we need to send a delimiter.
	useDelimiter := request.Execute == "guest-sync-delimited"

	return a.sendResponse(&qmpResponse, useDelimiter)
}

// sendResponse sends a response message.
func (a *Agent) sendResponse(resp *protocol.QMPResponse, useDelimiter bool) error {
	respData, err := json.Marshal(resp)
	if err != nil {
		logrus.WithError(err).WithField("response_id", resp.ID).Error("Failed to marshal response")
		// Send a fallback error if marshalling fails.
		fallbackResp := protocol.NewErrorResponse("InternalError", "Failed to marshal response")
		fallbackResp.ID = resp.ID
		fallbackData, _ := json.Marshal(fallbackResp)
		return a.commManager.SendResponse(fallbackData)
	}

	if useDelimiter {
		return a.commManager.SendDelimitedResponse(respData)
	}
	return a.commManager.SendResponse(respData)
}
