package protocol

import (
	"encoding/json"
)

// QMPRequest represents a QMP request
type QMPRequest struct {
	Execute   string      `json:"execute"`
	Arguments interface{} `json:"arguments,omitempty"`
	ID        interface{} `json:"id,omitempty"`
}

// QMPResponse represents a QMP response
type QMPResponse struct {
	Return interface{} `json:"return,omitempty"`
	Error  *QMPError   `json:"error,omitempty"`
	ID     interface{} `json:"id,omitempty"`
}

// QMPError represents a QMP error
type QMPError struct {
	Class string `json:"class"`
	Desc  string `json:"desc"`
}

// Command-specific argument structures
type PingArgs struct{}

type InfoArgs struct{}

type SyncArgs struct {
	ID int64 `json:"id"`
}

type ShutdownArgs struct {
	Mode string `json:"mode,omitempty"`
}

// Standard QEMU Guest Agent structures

// GuestAgentCommandInfo represents information about a guest agent command
type GuestAgentCommandInfo struct {
	Name            string `json:"name"`
	Enabled         bool   `json:"enabled"`
	SuccessResponse bool   `json:"success-response"`
}

// GuestAgentInfo represents information about the guest agent
type GuestAgentInfo struct {
	Version           string                  `json:"version"`
	SupportedCommands []GuestAgentCommandInfo `json:"supported_commands"`
}

// GuestOSInfo represents guest operating system information
type GuestOSInfo struct {
	KernelRelease string `json:"kernel-release,omitempty"`
	KernelVersion string `json:"kernel-version,omitempty"`
	Machine       string `json:"machine,omitempty"`
	ID            string `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	PrettyName    string `json:"pretty-name,omitempty"`
	Version       string `json:"version,omitempty"`
	VersionID     string `json:"version-id,omitempty"`
	Variant       string `json:"variant,omitempty"`
	VariantID     string `json:"variant-id,omitempty"`
}

// GuestHostName represents the guest hostname
type GuestHostName struct {
	HostName string `json:"host-name"`
}

// GuestUser represents a logged-in user
type GuestUser struct {
	User      string  `json:"user"`
	Domain    string  `json:"domain,omitempty"`
	LoginTime float64 `json:"login-time"`
}

// GuestTimezone represents timezone information
type GuestTimezone struct {
	Zone   string `json:"zone,omitempty"`
	Offset int    `json:"offset"`
}

// GuestIpAddressType represents IP address type
type GuestIpAddressType string

const (
	IPv4 GuestIpAddressType = "ipv4"
	IPv6 GuestIpAddressType = "ipv6"
)

// GuestIpAddress represents an IP address
type GuestIpAddress struct {
	IPAddress     string             `json:"ip-address"`
	IPAddressType GuestIpAddressType `json:"ip-address-type"`
	Prefix        int                `json:"prefix"`
}

// GuestNetworkInterfaceStat represents network interface statistics
type GuestNetworkInterfaceStat struct {
	RxBytes   int64 `json:"rx-bytes"`
	RxPackets int64 `json:"rx-packets"`
	RxErrs    int64 `json:"rx-errs"`
	RxDropped int64 `json:"rx-dropped"`
	TxBytes   int64 `json:"tx-bytes"`
	TxPackets int64 `json:"tx-packets"`
	TxErrs    int64 `json:"tx-errs"`
	TxDropped int64 `json:"tx-dropped"`
}

// GuestNetworkInterface represents a network interface
type GuestNetworkInterface struct {
	Name            string                     `json:"name"`
	HardwareAddress string                     `json:"hardware-address,omitempty"`
	IPAddresses     []GuestIpAddress           `json:"ip-addresses,omitempty"`
	Statistics      *GuestNetworkInterfaceStat `json:"statistics,omitempty"`
}

// GuestLogicalProcessor represents a logical processor
type GuestLogicalProcessor struct {
	LogicalID  int  `json:"logical-id"`
	Online     bool `json:"online"`
	CanOffline bool `json:"can-offline,omitempty"`
}

// GuestFsfreezeStatus represents filesystem freeze status
type GuestFsfreezeStatus string

const (
	FsfreezeStatusThawed GuestFsfreezeStatus = "thawed"
	FsfreezeStatusFrozen GuestFsfreezeStatus = "frozen"
)

// GuestFilesystemTrimResult represents filesystem trim result
type GuestFilesystemTrimResult struct {
	Path    string `json:"path"`
	Error   string `json:"error,omitempty"`
	Trimmed int64  `json:"trimmed,omitempty"`
	Minimum int64  `json:"minimum,omitempty"`
}

// GuestFilesystemTrimResponse represents filesystem trim response
type GuestFilesystemTrimResponse struct {
	Paths []GuestFilesystemTrimResult `json:"paths"`
}

// GuestDiskBusType represents disk bus type
type GuestDiskBusType string

const (
	DiskBusIDE     GuestDiskBusType = "ide"
	DiskBusFDC     GuestDiskBusType = "fdc"
	DiskBusSCSI    GuestDiskBusType = "scsi"
	DiskBusVirtio  GuestDiskBusType = "virtio"
	DiskBusXen     GuestDiskBusType = "xen"
	DiskBusUSB     GuestDiskBusType = "usb"
	DiskBusSATA    GuestDiskBusType = "sata"
	DiskBusSD      GuestDiskBusType = "sd"
	DiskBusUnknown GuestDiskBusType = "unknown"
)

// GuestPCIAddress represents a PCI address
type GuestPCIAddress struct {
	Domain   int `json:"domain"`
	Bus      int `json:"bus"`
	Slot     int `json:"slot"`
	Function int `json:"function"`
}

// GuestDiskAddress represents disk address information
type GuestDiskAddress struct {
	PCIController GuestPCIAddress  `json:"pci-controller"`
	BusType       GuestDiskBusType `json:"bus-type"`
	Bus           int              `json:"bus"`
	Target        int              `json:"target"`
	Unit          int              `json:"unit"`
	Serial        string           `json:"serial,omitempty"`
	Dev           string           `json:"dev,omitempty"`
}

// GuestFilesystemInfo represents filesystem information
type GuestFilesystemInfo struct {
	Name                 string             `json:"name"`
	Mountpoint           string             `json:"mountpoint"`
	Type                 string             `json:"type"`
	UsedBytes            int64              `json:"used-bytes,omitempty"`
	TotalBytes           int64              `json:"total-bytes,omitempty"`
	TotalBytesPrivileged int64              `json:"total-bytes-privileged,omitempty"`
	Disk                 []GuestDiskAddress `json:"disk"`
}

// GuestMemoryBlock represents a memory block
type GuestMemoryBlock struct {
	PhysIndex  int  `json:"phys-index"`
	Online     bool `json:"online"`
	CanOffline bool `json:"can-offline,omitempty"`
}

// GuestMemoryBlockInfo represents memory block information
type GuestMemoryBlockInfo struct {
	Size int64 `json:"size"`
}

// GuestDiskInfo represents disk information
type GuestDiskInfo struct {
	Name       string               `json:"name"`
	Partition  bool                 `json:"partition"`
	Address    *GuestDiskAddress    `json:"address,omitempty"`
	Alias      string               `json:"alias,omitempty"`
	HasMedia   bool                 `json:"has-media,omitempty"`
	Size       int64                `json:"size,omitempty"`
	Partitions []GuestPartitionInfo `json:"partitions,omitempty"`
}

// GuestPartitionInfo represents partition information
type GuestPartitionInfo struct {
	Number int    `json:"number"`
	Name   string `json:"name"`
	Size   int64  `json:"size,omitempty"`
}

// GuestSSHInfo represents SSH public key information
type GuestSSHInfo struct {
	Keys []string `json:"keys"`
}

// GuestSSHAddKeysArgs represents arguments for guest-ssh-add-authorized-keys command
type GuestSSHAddKeysArgs struct {
	Username string   `json:"username"`
	Keys     []string `json:"keys"`
}

// GuestSSHRemoveKeysArgs represents arguments for guest-ssh-remove-authorized-keys command
type GuestSSHRemoveKeysArgs struct {
	Username string   `json:"username"`
	Keys     []string `json:"keys"`
}

// GuestSSHGetKeysArgs represents arguments for guest-ssh-get-authorized-keys command
type GuestSSHGetKeysArgs struct {
	Username string `json:"username"`
}

// FstrimArgs represents arguments for fstrim command
type FstrimArgs struct {
	Minimum int64 `json:"minimum,omitempty"`
}

// FreezeListArgs represents arguments for fsfreeze-freeze-list command
type FreezeListArgs struct {
	Mountpoints []string `json:"mountpoints,omitempty"`
}

// GetTimeResponse represents the response for get-time command
type GetTimeResponse struct {
	Time int64 `json:"time"`
}

// SetTimeArgs represents arguments for set-time command
type SetTimeArgs struct {
	Time int64 `json:"time,omitempty"`
}

// EmptyResponse is used for commands that return no data, resulting in `{}`.
type EmptyResponse struct{}

// Helper function to create error response
func NewErrorResponse(class, desc string) *QMPResponse {
	return &QMPResponse{
		Error: &QMPError{
			Class: class,
			Desc:  desc,
		},
	}
}

// Helper function to create success response
func NewSuccessResponse(data interface{}) *QMPResponse {
	return &QMPResponse{
		Return: data,
	}
}

// Helper function to parse request arguments
func ParseArguments(args interface{}, target interface{}) error {
	if args == nil {
		return nil
	}

	data, err := json.Marshal(args)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}

// ParseRequest parses JSON data into QMPRequest
func ParseRequest(data []byte) (*QMPRequest, error) {
	var request QMPRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return nil, err
	}
	return &request, nil
}

// MarshalResponse marshals QMPResponse into JSON data
func MarshalResponse(response *QMPResponse) ([]byte, error) {
	return json.Marshal(response)
}
