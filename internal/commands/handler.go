// Package commands contains the implementation for all guest agent commands.
package commands

import (
	"encoding/json"
	"fmt"
	"mac-guest-agent/internal/protocol"

	log "github.com/sirupsen/logrus"
)

// Command defines a standard structure for a guest agent command, inspired by
// the official QEMU Guest Agent implementation.
type Command struct {
	// Name is the public name of the command (e.g., "guest-info").
	Name string

	// Handler is the function that executes the command's logic.
	// It takes a json.RawMessage as input for arguments and returns the result
	// or an error.
	Handler func(req json.RawMessage) (interface{}, error)

	// Enabled controls whether the command is active and can be executed.
	Enabled bool
}

// CommandRegistry holds all registered commands for the agent.
// The key is the command name.
var CommandRegistry = make(map[string]*Command)

// RegisterCommand adds a new command to the registry.
// If the command is nil or its name is empty, it will not be registered.
func RegisterCommand(cmd *Command) {
	if cmd == nil || cmd.Name == "" {
		return
	}
	log.Debugf("Registering command: %s", cmd.Name)
	CommandRegistry[cmd.Name] = cmd
}

// HandleCommand processes an incoming command request using the CommandRegistry.
func HandleCommand(req protocol.QMPRequest) protocol.QMPResponse {
	// 对于高频心跳命令使用Debug级别，其他命令使用Info级别
	if req.Execute == "guest-ping" || req.Execute == "guest-sync" || req.Execute == "guest-sync-delimited" {
		log.Debugf("Handling command: %s", req.Execute)
	} else {
		log.Infof("Handling command: %s", req.Execute)
	}

	cmd, ok := CommandRegistry[req.Execute]
	if !ok || !cmd.Enabled {
		log.Errorf("Command not found or disabled: %s", req.Execute)
		return protocol.QMPResponse{
			Error: &protocol.QMPError{
				Class: "CommandNotFound",
				Desc:  fmt.Sprintf("The command %s has not been found", req.Execute),
			},
		}
	}

	// The arguments in QMPRequest are an interface{}, but our handlers expect
	// json.RawMessage. We need to marshal it back to raw JSON.
	var argsJSON json.RawMessage
	if req.Arguments != nil {
		// If arguments are already in raw message form, use them directly.
		if raw, ok := req.Arguments.(json.RawMessage); ok {
			argsJSON = raw
		} else {
			argBytes, err := json.Marshal(req.Arguments)
			if err != nil {
				log.Errorf("Failed to marshal arguments for command %s: %v", req.Execute, err)
				return protocol.QMPResponse{
					Error: &protocol.QMPError{Class: "InvalidParameter", Desc: "could not marshal arguments"},
				}
			}
			argsJSON = argBytes
		}
	}

	// Execute the command handler.
	result, err := cmd.Handler(argsJSON)
	if err != nil {
		log.Errorf("Error executing command %s: %v", req.Execute, err)
		return protocol.QMPResponse{
			Error: &protocol.QMPError{
				Class: "GenericError",
				Desc:  err.Error(),
			},
		}
	}

	return protocol.QMPResponse{
		Return: result,
	}
}
