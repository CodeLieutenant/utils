package utils

import (
	"bytes"
	"net"
	"sync"
)

const (
	UnknownIP           = "UNKNOWN IP"
	HeaderXForwardedFor = "X-Forwarded-For"
	HeaderXRealIP       = "X-Real-IP"
)

var (
	ip      = ""
	ips     = make([]string, 0, 5)
	onceIP  = &sync.Once{}
	onceIPs = &sync.Once{}
)

type Peekable interface {
	Peek(key string) []byte
}

// GetLocalIP Returns IP address of local machine, empty string if fails
func GetLocalIP() string {
	onceIP.Do(func() {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			ip = ""
			return
		}

		for _, address := range addrs {
			ipnet, ok := address.(*net.IPNet)

			// check the address type and if it is not a loopback the display it
			if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
				break
			}
		}
	})

	return ip
}

// Returns strinngs slice of IP found on local machine
func GetLocalIPs() []string {
	onceIPs.Do(func() {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			ips = nil
			return
		}

		for _, address := range addrs {
			ipnet, ok := address.(*net.IPNet)

			// check the address type and if it is not a loopback the display it
			if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	})

	return ips
}

// RealIp - Extracts first ip address from Peekable interface seperated by coma
// Returns nil if no values are presemt
func RealIP(peekable Peekable) []byte {
	ipHeader := peekable.Peek(HeaderXForwardedFor)

	if len(ipHeader) == 0 {
		ipHeader = peekable.Peek(HeaderXRealIP)
	}

	if len(ipHeader) == 0 {
		return nil
	}

	firstIndex := bytes.IndexRune(ipHeader, ',')

	ip := ipHeader

	if firstIndex != -1 {
		ip = ipHeader[:firstIndex]
	}

	return ip
}
