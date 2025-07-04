package commands

import (
	"bufio"
	"encoding/json"
	"mac-guest-agent/internal/protocol"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterCommand(&Command{
		Name:    "guest-get-users",
		Handler: handleGetUsers,
		Enabled: true,
	})
}

// handleGetUsers handles the guest-get-users command.
func handleGetUsers(req json.RawMessage) (interface{}, error) {
	users, err := getLoggedInUsers()
	if err != nil {
		logrus.WithError(err).Error("Failed to get logged-in users")
		return nil, err
	}
	logrus.WithField("user_count", len(users)).Info("Successfully retrieved logged-in users")
	return users, nil
}

// getLoggedInUsers retrieves the currently logged-in users.
// It uses the `who` command, which is standard and reliable.
func getLoggedInUsers() ([]protocol.GuestUser, error) {
	cmd := exec.Command("who")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	userMap := make(map[string]protocol.GuestUser)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		user := parseWhoLine(line)
		if user == nil {
			continue
		}

		// Keep the user with the earliest login time.
		existing, exists := userMap[user.User]
		if !exists || user.LoginTime < existing.LoginTime {
			userMap[user.User] = *user
		}
	}

	users := make([]protocol.GuestUser, 0, len(userMap))
	for _, user := range userMap {
		users = append(users, user)
	}

	return users, nil
}

// parseWhoLine parses a line from the output of the `who` command.
func parseWhoLine(line string) *protocol.GuestUser {
	// `who` output format: username terminal date time
	// Example: user1 console  Jun 29 12:00
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return nil
	}

	username := fields[0]
	// The date is typically in "Mmm dd HH:MM" format.
	timeStr := strings.Join(fields[2:5], " ")

	// The `who` command doesn't provide a year, so we assume the current year.
	// This is a limitation of the command itself.
	loginTime, err := time.Parse("Jan _2 15:04", timeStr)
	if err != nil {
		// If parsing fails, use the current time as a fallback.
		loginTime = time.Now()
	} else {
		now := time.Now()
		loginTime = loginTime.AddDate(now.Year(), 0, 0)
		// If the parsed time is in the future, it must be from the previous year.
		if loginTime.After(now) {
			loginTime = loginTime.AddDate(-1, 0, 0)
		}
	}

	return &protocol.GuestUser{
		User:      username,
		LoginTime: float64(loginTime.Unix()),
		// Domain is not available in `who` output on macOS.
		Domain: "",
	}
}
