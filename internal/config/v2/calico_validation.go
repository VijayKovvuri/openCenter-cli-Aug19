package v2

import (
	"fmt"
	"net"
	"strings"
)

const (
	CalicoInterfaceAutodetectFirstFound = "first-found"
	CalicoInterfaceAutodetectInterface  = "interface"
	CalicoInterfaceAutodetectCIDR       = "cidr"
)

// ResolveCalicoInterfaceAutodetect returns the effective Calico node address
// autodetection mode without modifying the configuration. An omitted mode uses
// Calico's first-found behavior. Only fields selected by the mode are checked;
// values left over from another mode are intentionally ignored.
func ResolveCalicoInterfaceAutodetect(config *CalicoConfig) (string, error) {
	if config == nil {
		return CalicoInterfaceAutodetectFirstFound, nil
	}

	mode := strings.ToLower(strings.TrimSpace(config.CalicoInterfaceAutodetect))
	switch mode {
	case "":
		return CalicoInterfaceAutodetectFirstFound, nil
	case CalicoInterfaceAutodetectFirstFound:
		return CalicoInterfaceAutodetectFirstFound, nil
	case CalicoInterfaceAutodetectInterface:
		if strings.TrimSpace(config.CNIIface) == "" {
			return "", fmt.Errorf("cni_iface is required when calico_interface_autodetect is %q", CalicoInterfaceAutodetectInterface)
		}
		return CalicoInterfaceAutodetectInterface, nil
	case CalicoInterfaceAutodetectCIDR:
		if !validCalicoIPv4CIDR(config.AutodetectCIDR) {
			return "", fmt.Errorf("autodetect_cidr must be a valid IPv4 CIDR when calico_interface_autodetect is %q", CalicoInterfaceAutodetectCIDR)
		}
		return CalicoInterfaceAutodetectCIDR, nil
	default:
		return "", fmt.Errorf("calico_interface_autodetect must be one of %q, %q, or %q", CalicoInterfaceAutodetectFirstFound, CalicoInterfaceAutodetectInterface, CalicoInterfaceAutodetectCIDR)
	}
}

func validCalicoIPv4CIDR(value string) bool {
	ip, network, err := net.ParseCIDR(strings.TrimSpace(value))
	if err != nil || ip == nil || network == nil {
		return false
	}
	_, bits := network.Mask.Size()
	return bits == 32 && ip.To4() != nil
}
