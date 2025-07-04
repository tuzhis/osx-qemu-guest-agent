package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"mac-guest-agent/internal/protocol"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ShutdownArgs defines the arguments for the guest-shutdown command.
type ShutdownArgs struct {
	Mode string `json:"mode"`
}

func init() {
	RegisterCommand(&Command{
		Name:    "guest-shutdown",
		Handler: handleGuestShutdown,
		Enabled: true,
	})
}

// handleGuestShutdown handles the guest-shutdown command.
func handleGuestShutdown(req json.RawMessage) (interface{}, error) {
	args := ShutdownArgs{Mode: "powerdown"} // Default mode
	if len(req) > 0 && string(req) != "null" {
		if err := json.Unmarshal(req, &args); err != nil {
			return nil, fmt.Errorf("failed to parse arguments: %v", err)
		}
	}

	logrus.WithField("mode", args.Mode).Info("Received shutdown command")

	// Execute the shutdown command in the background to avoid blocking the response.
	go func() {
		// A short delay to ensure the response can be sent.
		time.Sleep(200 * time.Millisecond)

		switch args.Mode {
		case "powerdown", "halt": // halt and powerdown use the same implementation.
			executePowerDown()
		case "reboot":
			executeReboot()
		default:
			logrus.WithField("mode", args.Mode).Error("Unsupported shutdown mode")
		}
	}()

	// Return success immediately.
	return protocol.EmptyResponse{}, nil
}

// executePowerDown performs the powerdown operation.
func executePowerDown() {
	logrus.Info("Executing powerdown with 10s timeout...")

	// Create a context with a timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start the graceful shutdown process.
	gracefulDone := make(chan bool, 1)
	go func() {
		// Clean up application states and restore settings.
		clearAllApplicationStates()

		// Use the fastest shutdown method.
		performImmediateShutdown()
		gracefulDone <- true
	}()

	// Start a timer for forced shutdown.
	forceTimer := time.NewTimer(10 * time.Second)
	defer forceTimer.Stop()

	go func() {
		<-forceTimer.C
		logrus.Warning("Graceful shutdown timeout after 10s, forcing immediate powerdown...")
		performForceShutdown()
	}()

	// Wait for graceful shutdown to complete or for the timeout.
	select {
	case <-gracefulDone:
		logrus.Info("Graceful powerdown completed")
	case <-ctx.Done():
		logrus.Warning("Powerdown context timeout")
	}
}

// executeReboot performs the reboot operation.
func executeReboot() {
	logrus.Info("Executing reboot with 10s timeout...")

	// Create a context with a timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start the graceful reboot process.
	gracefulDone := make(chan bool, 1)
	go func() {
		// Clean up application states and restore settings.
		clearAllApplicationStates()

		// Use the fastest reboot method.
		performImmediateReboot()
		gracefulDone <- true
	}()

	// Start a timer for forced reboot.
	forceTimer := time.NewTimer(10 * time.Second)
	defer forceTimer.Stop()

	go func() {
		<-forceTimer.C
		logrus.Warning("Graceful reboot timeout after 10s, forcing immediate reboot...")
		performForceReboot()
	}()

	// Wait for graceful reboot to complete or for the timeout.
	select {
	case <-gracefulDone:
		logrus.Info("Graceful reboot completed")
	case <-ctx.Done():
		logrus.Warning("Reboot context timeout")
	}
}

// clearAllApplicationStates cleans up application states and restores settings.
func clearAllApplicationStates() {
	logrus.Info("Starting graceful application state cleanup...")

	// 1. Disable system resume features immediately.
	disableAllResumeFeatures()

	// 2. Gracefully close user applications, allowing them to save.
	gracefullyCloseUserApplications()

	// 3. Clean all saved state files.
	removeAllSavedStates()

	// 4. Clean system resume-related files.
	clearSystemResumeFiles()

	logrus.Info("Application state cleanup completed")
}

// disableAllResumeFeatures disables all resume features.
func disableAllResumeFeatures() {
	logrus.Info("Disabling all resume features...")

	// System-level settings
	systemCommands := [][]string{
		{"defaults", "write", "com.apple.loginwindow", "TALLogoutSavesState", "-bool", "false"},
		{"defaults", "write", "com.apple.loginwindow", "LoginwindowLaunchesRelaunchApps", "-bool", "false"},
		{"defaults", "write", "-g", "NSQuitAlwaysKeepsWindows", "-bool", "false"},
		{"defaults", "delete", "com.apple.loginwindow", "RestoreWindowState"},
		{"defaults", "write", "com.apple.loginwindow", "AutolaunchedApplicationDictionary", "-array"},
	}

	for _, cmdArgs := range systemCommands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		err := cmd.Run()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"command": cmdArgs,
				"error":   err,
			}).Debug("System command failed (may be normal)")
		}
	}

	// Immediately sync settings to disk.
	exec.Command("sync").Run()
}

// gracefullyCloseUserApplications gracefully closes user applications, allowing them to save.
func gracefullyCloseUserApplications() {
	logrus.Info("Gracefully closing user applications with concurrent processing...")

	// Get the list of all running applications.
	appList := getRunningApplications()
	if len(appList) == 0 {
		logrus.Info("No user applications to close")
		return
	}

	logrus.WithField("app_count", len(appList)).Info("Found applications to close")

	// Use multithreading for concurrent application closure.
	var wg sync.WaitGroup

	// Start a goroutine for each application, including Finder.
	for _, appName := range appList {
		if appName == "System Events" || appName == "loginwindow" {
			continue // Skip core system processes.
		}

		wg.Add(1)
		go func(app string) {
			defer wg.Done()
			if app == "Finder" {
				closeFinderConcurrently(app)
			} else {
				closeApplicationConcurrently(app)
			}
		}(appName)
	}

	// Wait for all applications to close, or for a timeout.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Set a 7-second timeout to avoid indefinite waiting.
	select {
	case <-done:
		logrus.Info("All applications (including Finder) closed successfully")
	case <-time.After(7 * time.Second):
		logrus.Warning("Application closure timeout after 7s, some apps may still be running")
	}
}

// removeAllSavedStates removes all saved application state files.
func removeAllSavedStates() {
	logrus.Info("Removing all saved application states...")

	// List of cleanup commands.
	cleanupCommands := []string{
		"rm -rf ~/Library/Saved\\ Application\\ State/*.savedState 2>/dev/null || true",
		"rm -rf ~/Library/Saved\\ Application\\ State/* 2>/dev/null || true",
		"rm -rf ~/Library/Preferences/ByHost/com.apple.loginwindow.* 2>/dev/null || true",
		"rm -rf ~/Library/Preferences/com.apple.loginwindow.plist 2>/dev/null || true",
		"rm -rf ~/Library/Application\\ Support/CrashReporter/* 2>/dev/null || true",
		"rm -rf ~/Library/Caches/* 2>/dev/null || true",
	}

	for _, cmdStr := range cleanupCommands {
		cmd := exec.Command("sh", "-c", cmdStr)
		err := cmd.Run()
		if err != nil {
			logrus.WithField("command", cmdStr).WithError(err).Warn("Cleanup command failed")
		}
	}
}

// clearSystemResumeFiles clears system resume-related files.
func clearSystemResumeFiles() {
	logrus.Info("Clearing system resume files...")
	cmd := exec.Command("rm", "-rf", "/var/vm/sleepimage")
	err := cmd.Run()
	if err != nil {
		logrus.WithError(err).Warn("Failed to remove sleepimage")
	}
}

// performImmediateShutdown performs an immediate shutdown.
func performImmediateShutdown() {
	logrus.Info("Performing immediate shutdown via osascript...")
	script := "tell app \"System Events\" to shut down"
	cmd := exec.Command("osascript", "-e", script)
	err := cmd.Run()
	if err != nil {
		logrus.WithError(err).Error("osascript shutdown failed, trying fallback")
		performForceShutdown()
	}
}

// performImmediateReboot performs an immediate reboot.
func performImmediateReboot() {
	logrus.Info("Performing immediate reboot via osascript...")
	script := "tell app \"System Events\" to restart"
	cmd := exec.Command("osascript", "-e", script)
	err := cmd.Run()
	if err != nil {
		logrus.WithError(err).Error("osascript reboot failed, trying fallback")
		performForceReboot()
	}
}

// performForceShutdown performs a forced shutdown.
func performForceShutdown() {
	logrus.Warning("Performing force shutdown via 'shutdown -h now'...")
	cmd := exec.Command("shutdown", "-h", "now")
	err := cmd.Run()
	if err != nil {
		logrus.WithError(err).Error("Force shutdown command failed")
	}
}

// performForceReboot performs a forced reboot.
func performForceReboot() {
	logrus.Warning("Performing force reboot via 'shutdown -r now'...")
	cmd := exec.Command("shutdown", "-r", "now")
	err := cmd.Run()
	if err != nil {
		logrus.WithError(err).Error("Force reboot command failed")
	}
}

// getRunningApplications gets the list of running applications.
func getRunningApplications() []string {
	script := "tell application \"System Events\" to get name of every process whose background only is false"
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		logrus.WithError(err).Error("Failed to get running applications")
		return nil
	}
	apps := strings.Split(string(out), ", ")
	for i, app := range apps {
		apps[i] = strings.TrimSpace(app)
	}
	return apps
}

// closeApplicationConcurrently closes an application concurrently.
func closeApplicationConcurrently(appName string) {
	logrus.WithField("app", appName).Info("Attempting to close application")
	if !gracefulQuitApplication(appName) {
		logrus.WithField("app", appName).Warning("Graceful quit failed, forcing quit")
		if !forceQuitApplication(appName) {
			logrus.WithField("app", appName).Error("Force quit failed, killing process")
			killApplication(appName)
		}
	}
}

// gracefulQuitApplication gracefully quits an application.
func gracefulQuitApplication(appName string) bool {
	script := fmt.Sprintf("quit app \"%s\"", appName)
	cmd := exec.Command("osascript", "-e", script)
	err := cmd.Run()
	return err == nil
}

// forceQuitApplication forcefully quits an application.
func forceQuitApplication(appName string) bool {
	script := fmt.Sprintf("tell application \"System Events\" to unix id of process \"%s\"", appName)
	cmd := exec.Command("osascript", "-e", script)
	pid, err := cmd.Output()
	if err != nil {
		return false
	}
	pidStr := strings.TrimSpace(string(pid))
	killCmd := exec.Command("kill", "-9", pidStr)
	return killCmd.Run() == nil
}

// killApplication kills an application process.
func killApplication(appName string) {
	killCmd := exec.Command("killall", appName)
	killCmd.Run()
}

// closeFinderConcurrently closes the Finder application concurrently.
func closeFinderConcurrently(appName string) {
	logrus.Info("Attempting to close Finder")
	if !gracefulQuitFinder() {
		logrus.Warning("Graceful quit for Finder failed, forcing quit")
		if !forceQuitFinder() {
			logrus.Error("Force quit for Finder failed, killing process")
			killFinder()
		}
	}
}

// gracefulQuitFinder gracefully quits the Finder.
func gracefulQuitFinder() bool {
	script := "tell application \"Finder\" to quit"
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run() == nil
}

// forceQuitFinder forcefully quits the Finder.
func forceQuitFinder() bool {
	script := "tell application \"System Events\" to unix id of process \"Finder\""
	cmd := exec.Command("osascript", "-e", script)
	pid, err := cmd.Output()
	if err != nil {
		return false
	}
	pidStr := strings.TrimSpace(string(pid))
	killCmd := exec.Command("kill", "-9", pidStr)
	return killCmd.Run() == nil
}

// killFinder kills the Finder process.
func killFinder() {
	killCmd := exec.Command("killall", "Finder")
	killCmd.Run()
}
