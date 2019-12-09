//Copyright 2019, Intel Corporation

package rwogluster

import (
	"errors"
	"net"
)

func resolveLocalIP() (string, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, address := range addresses {
		if ip, ok := address.(*net.IPNet); ok && !ip.IP.IsLoopback() && ip.IP.To4() != nil {
			return ip.IP.String(), nil
		}
	}

	return "", errors.New("No network found to resolve IP")
}
