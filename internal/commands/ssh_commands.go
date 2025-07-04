package commands

import (
	"encoding/json"
	"fmt"
	"mac-guest-agent/internal/protocol"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterCommand(&Command{
		Name:    "guest-ssh-get-authorized-keys",
		Handler: handleSSHGetAuthorizedKeys,
		Enabled: true,
	})
	RegisterCommand(&Command{
		Name:    "guest-ssh-add-authorized-keys",
		Handler: handleSSHAddAuthorizedKeys,
		Enabled: true,
	})
	RegisterCommand(&Command{
		Name:    "guest-ssh-remove-authorized-keys",
		Handler: handleSSHRemoveAuthorizedKeys,
		Enabled: true,
	})
}

// handleSSHGetAuthorizedKeys handles the guest-ssh-get-authorized-keys command.
func handleSSHGetAuthorizedKeys(req json.RawMessage) (interface{}, error) {
	var args protocol.GuestSSHGetKeysArgs
	if err := json.Unmarshal(req, &args); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %v", err)
	}

	logrus.WithField("username", args.Username).Info("SSH get authorized keys requested")

	// 由于安全原因，macOS版本的guest-agent不支持SSH密钥管理
	// 返回一个安全错误，而不是实际执行操作
	return nil, fmt.Errorf("SSH key management is not supported in macOS Guest Agent for security reasons")
}

// handleSSHAddAuthorizedKeys handles the guest-ssh-add-authorized-keys command.
func handleSSHAddAuthorizedKeys(req json.RawMessage) (interface{}, error) {
	var args protocol.GuestSSHAddKeysArgs
	if err := json.Unmarshal(req, &args); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"username":  args.Username,
		"key_count": len(args.Keys),
	}).Info("SSH add authorized keys requested")

	// 由于安全原因，macOS版本的guest-agent不支持SSH密钥管理
	// 返回一个安全错误，而不是实际执行操作
	return nil, fmt.Errorf("SSH key management is not supported in macOS Guest Agent for security reasons")
}

// handleSSHRemoveAuthorizedKeys handles the guest-ssh-remove-authorized-keys command.
func handleSSHRemoveAuthorizedKeys(req json.RawMessage) (interface{}, error) {
	var args protocol.GuestSSHRemoveKeysArgs
	if err := json.Unmarshal(req, &args); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"username":  args.Username,
		"key_count": len(args.Keys),
	}).Info("SSH remove authorized keys requested")

	// 由于安全原因，macOS版本的guest-agent不支持SSH密钥管理
	// 返回一个安全错误，而不是实际执行操作
	return nil, fmt.Errorf("SSH key management is not supported in macOS Guest Agent for security reasons")
}

// 以下是实际实现SSH密钥管理的函数，但在macOS版本中不会被调用
// 保留这些代码是为了未来可能的功能扩展

// getAuthorizedKeysPath returns the path to the authorized_keys file for a user.
func getAuthorizedKeysPath(username string) (string, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return "", fmt.Errorf("user not found: %v", err)
	}

	sshDir := filepath.Join(u.HomeDir, ".ssh")
	return filepath.Join(sshDir, "authorized_keys"), nil
}

// getAuthorizedKeys reads the authorized_keys file for a user.
func getAuthorizedKeys(username string) ([]string, error) {
	path, err := getAuthorizedKeysPath(username)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	lines := []string{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}

	return lines, nil
}

// addAuthorizedKeys adds keys to the authorized_keys file for a user.
func addAuthorizedKeys(username string, keys []string) error {
	path, err := getAuthorizedKeysPath(username)
	if err != nil {
		return err
	}

	// Create .ssh directory if it doesn't exist
	sshDir := filepath.Dir(path)
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return err
	}

	// Read existing keys
	existingKeys, err := getAuthorizedKeys(username)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Add new keys, avoiding duplicates
	keyMap := make(map[string]bool)
	for _, key := range existingKeys {
		keyMap[key] = true
	}

	var newKeys []string
	for _, key := range keys {
		if !keyMap[key] {
			newKeys = append(newKeys, key)
			keyMap[key] = true
		}
	}

	// Combine all keys
	allKeys := append(existingKeys, newKeys...)

	// Write back to file
	data := []byte(strings.Join(allKeys, "\n") + "\n")
	return os.WriteFile(path, data, 0600)
}

// removeAuthorizedKeys removes keys from the authorized_keys file for a user.
func removeAuthorizedKeys(username string, keys []string) error {
	path, err := getAuthorizedKeysPath(username)
	if err != nil {
		return err
	}

	// Read existing keys
	existingKeys, err := getAuthorizedKeys(username)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Create a map of keys to remove
	removeMap := make(map[string]bool)
	for _, key := range keys {
		removeMap[key] = true
	}

	// Filter out keys to remove
	var remainingKeys []string
	for _, key := range existingKeys {
		if !removeMap[key] {
			remainingKeys = append(remainingKeys, key)
		}
	}

	// Write back to file
	data := []byte(strings.Join(remainingKeys, "\n"))
	if len(remainingKeys) > 0 {
		data = append(data, '\n')
	}
	return os.WriteFile(path, data, 0600)
}
