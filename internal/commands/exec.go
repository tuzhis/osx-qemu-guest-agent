package commands

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ExecProcess represents a process executed via guest-exec.
type ExecProcess struct {
	PID      int       `json:"pid"`
	Exited   bool      `json:"exited"`
	ExitCode int       `json:"exitcode"`
	OutData  string    `json:"out-data,omitempty"`
	ErrData  string    `json:"err-data,omitempty"`
	Time     time.Time `json:"-"` // Internal use only
}

// GuestExecArgs represents the arguments for the guest-exec command.
type GuestExecArgs struct {
	Path          string   `json:"path"`
	Arg           []string `json:"arg,omitempty"`
	Env           []string `json:"env,omitempty"`
	InputData     string   `json:"input-data,omitempty"`
	CaptureOutput bool     `json:"capture-output,omitempty"`
}

// GuestExecStatusArgs represents the arguments for the guest-exec-status command.
type GuestExecStatusArgs struct {
	PID int `json:"pid"`
}

// GuestExecResponse represents the response for the guest-exec command.
type GuestExecResponse struct {
	PID int `json:"pid"`
}

var (
	execProcesses      = make(map[int]*ExecProcess)
	execProcessesMutex sync.Mutex
	nextPID            = 1
)

func init() {
	RegisterCommand(&Command{
		Name:    "guest-exec",
		Handler: handleGuestExec,
		Enabled: true,
	})
	RegisterCommand(&Command{
		Name:    "guest-exec-status",
		Handler: handleGuestExecStatus,
		Enabled: true,
	})
}

// handleGuestExec handles the guest-exec command.
func handleGuestExec(req json.RawMessage) (interface{}, error) {
	var args GuestExecArgs
	if err := json.Unmarshal(req, &args); err != nil {
		return nil, fmt.Errorf("failed to parse arguments for guest-exec: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"path": args.Path,
		"args": args.Arg,
	}).Info("Guest exec command requested")

	// 由于安全原因，macOS版本的guest-agent不支持执行命令
	// 返回一个安全错误，而不是实际执行命令
	return nil, fmt.Errorf("guest-exec is not supported in macOS Guest Agent for security reasons")
}

// handleGuestExecStatus handles the guest-exec-status command.
func handleGuestExecStatus(req json.RawMessage) (interface{}, error) {
	var args GuestExecStatusArgs
	if err := json.Unmarshal(req, &args); err != nil {
		return nil, fmt.Errorf("failed to parse arguments for guest-exec-status: %v", err)
	}

	logrus.WithField("pid", args.PID).Info("Guest exec status requested")

	// 由于我们不支持guest-exec，所以这里也返回一个错误
	return nil, fmt.Errorf("guest-exec is not supported in macOS Guest Agent")
}

// 以下是实际执行命令的函数，但在macOS版本中不会被调用
// 保留这些代码是为了未来可能的功能扩展
func executeCommand(args GuestExecArgs) (*ExecProcess, error) {
	cmd := exec.Command(args.Path, args.Arg...)
	cmd.Env = args.Env

	var outData, errData []byte
	var err error

	if args.CaptureOutput {
		outData, err = cmd.Output()
	} else {
		err = cmd.Run()
	}

	execProcessesMutex.Lock()
	defer execProcessesMutex.Unlock()

	pid := nextPID
	nextPID++

	process := &ExecProcess{
		PID:      pid,
		Exited:   true,
		ExitCode: 0,
		Time:     time.Now(),
	}

	if args.CaptureOutput {
		process.OutData = string(outData)
		process.ErrData = string(errData)
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			process.ExitCode = exitError.ExitCode()
		} else {
			process.ExitCode = 1
			process.ErrData = err.Error()
		}
	}

	execProcesses[pid] = process
	return process, nil
}

// 清理过期的进程记录
func cleanupOldProcesses() {
	execProcessesMutex.Lock()
	defer execProcessesMutex.Unlock()

	now := time.Now()
	for pid, process := range execProcesses {
		// 清理超过30分钟的进程记录
		if now.Sub(process.Time) > 30*time.Minute {
			delete(execProcesses, pid)
		}
	}
}
