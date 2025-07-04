package commands

import (
	"bufio"
	"encoding/json"
	"mac-guest-agent/internal/protocol"
	"net"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func init() {
	RegisterCommand(&Command{
		Name:    "guest-network-get-interfaces",
		Handler: handleNetworkGetInterfaces,
		Enabled: true,
	})
}

// handleNetworkGetInterfaces handles the guest-network-get-interfaces command.
func handleNetworkGetInterfaces(req json.RawMessage) (interface{}, error) {
	interfaces, err := getNetworkInterfaces()
	if err != nil {
		logrus.WithError(err).Error("Failed to get network interfaces")
		return nil, err
	}

	logrus.WithField("interface_count", len(interfaces)).Info("Successfully retrieved network interfaces")
	return interfaces, nil
}

// getNetworkInterfaces retrieves network interface information.
func getNetworkInterfaces() ([]protocol.GuestNetworkInterface, error) {
	var interfaces []protocol.GuestNetworkInterface

	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range netInterfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		guestIface := protocol.GuestNetworkInterface{
			Name: iface.Name,
		}

		if iface.HardwareAddr != nil {
			guestIface.HardwareAddress = iface.HardwareAddr.String()
		}

		addrs, err := iface.Addrs()
		if err == nil {
			var ipAddresses []protocol.GuestIpAddress
			for _, addr := range addrs {
				if ipAddr := parseIPAddress(addr); ipAddr != nil {
					ipAddresses = append(ipAddresses, *ipAddr)
				}
			}
			guestIface.IPAddresses = ipAddresses
		}

		if stats := getInterfaceStatistics(iface.Name); stats != nil {
			guestIface.Statistics = stats
		}

		interfaces = append(interfaces, guestIface)
	}

	return interfaces, nil
}

// parseIPAddress parses an IP address from a net.Addr.
func parseIPAddress(addr net.Addr) *protocol.GuestIpAddress {
	var ip net.IP
	var mask net.IPMask
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
		mask = v.Mask
	case *net.IPAddr:
		ip = v.IP
	}

	if ip == nil {
		return nil
	}

	var ipType protocol.GuestIpAddressType
	if ip.To4() != nil {
		ipType = protocol.IPv4
	} else {
		ipType = protocol.IPv6
	}

	prefix, _ := mask.Size()

	return &protocol.GuestIpAddress{
		IPAddress:     ip.String(),
		IPAddressType: ipType,
		Prefix:        prefix,
	}
}

// getInterfaceStatistics retrieves statistics for a given interface.
func getInterfaceStatistics(ifaceName string) *protocol.GuestNetworkInterfaceStat {
	cmd := exec.Command("netstat", "-ibn")
	output, err := cmd.Output()
	if err != nil {
		logrus.WithError(err).WithField("interface", ifaceName).Debug("Failed to get network statistics")
		return nil
	}
	return parseNetstatOutput(string(output), ifaceName)
}

// parseNetstatOutput parses the output of the netstat command.
func parseNetstatOutput(output, ifaceName string) *protocol.GuestNetworkInterfaceStat {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) < 10 || fields[0] != ifaceName {
			continue
		}

		stats := &protocol.GuestNetworkInterfaceStat{}
		if val, err := strconv.ParseInt(fields[4], 10, 64); err == nil {
			stats.RxPackets = val
		}
		if val, err := strconv.ParseInt(fields[5], 10, 64); err == nil {
			stats.RxErrs = val
		}
		if val, err := strconv.ParseInt(fields[6], 10, 64); err == nil {
			stats.RxBytes = val
		}
		if val, err := strconv.ParseInt(fields[7], 10, 64); err == nil {
			stats.TxPackets = val
		}
		if val, err := strconv.ParseInt(fields[8], 10, 64); err == nil {
			stats.TxErrs = val
		}
		if val, err := strconv.ParseInt(fields[9], 10, 64); err == nil {
			stats.TxBytes = val
		}
		return stats
	}
	return nil
}
