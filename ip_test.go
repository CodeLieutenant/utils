package utils_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CodeLieutenant/utils"
)

func TestGetLocalIP(t *testing.T) {
	t.Parallel()
	assert := require.New(t)
	addrs, err := net.InterfaceAddrs()
	assert.NoError(err)
	var ip string

	for _, address := range addrs {
		ipnet, ok := address.(*net.IPNet)

		// check the address type and if it is not a loopback the display it
		if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ip = ipnet.IP.String()
			break
		}
	}

	localIP := utils.GetLocalIP()

	assert.Equal(ip, localIP)

	// Testing the cached version
	localIP = utils.GetLocalIP()

	assert.Equal(ip, localIP)
}

func TestGetLocalIPs(t *testing.T) {
	t.Parallel()
	assert := require.New(t)
	addrs, err := net.InterfaceAddrs()
	assert.NoError(err)
	var ips []string

	for _, address := range addrs {
		ipnet, ok := address.(*net.IPNet)

		// check the address type and if it is not a loopback the display it
		if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ips = append(ips, ipnet.IP.String())
		}
	}

	localIps := utils.GetLocalIPs()

	assert.EqualValues(ips, localIps)

	// Testing the cached version
	localIps = utils.GetLocalIPs()

	assert.EqualValues(ips, localIps)
}

type mockPeekable struct {
	headers map[string][]byte
}

func (m *mockPeekable) Peek(key string) []byte {
	return m.headers[key]
}

func TestRealIP(t *testing.T) {
	t.Parallel()
	req := require.New(t)

	testCases := []struct {
		name           string
		headers        map[string][]byte
		expectedResult []byte
	}{
		{
			name: "X-Forwarded-For header with single IP",
			headers: map[string][]byte{
				utils.HeaderXForwardedFor: []byte("192.168.1.1"),
			},
			expectedResult: []byte("192.168.1.1"),
		},
		{
			name: "X-Forwarded-For header with multiple IPs",
			headers: map[string][]byte{
				utils.HeaderXForwardedFor: []byte("192.168.1.1,10.0.0.1,172.16.0.1"),
			},
			expectedResult: []byte("192.168.1.1"),
		},
		{
			name: "X-Real-IP header when X-Forwarded-For is empty",
			headers: map[string][]byte{
				utils.HeaderXRealIP: []byte("10.0.0.1"),
			},
			expectedResult: []byte("10.0.0.1"),
		},
		{
			name: "X-Forwarded-For takes precedence over X-Real-IP",
			headers: map[string][]byte{
				utils.HeaderXForwardedFor: []byte("192.168.1.1"),
				utils.HeaderXRealIP:       []byte("10.0.0.1"),
			},
			expectedResult: []byte("192.168.1.1"),
		},
		{
			name: "X-Forwarded-For with spaces around comma",
			headers: map[string][]byte{
				utils.HeaderXForwardedFor: []byte("192.168.1.1, 10.0.0.1"),
			},
			expectedResult: []byte("192.168.1.1"),
		},
		{
			name:           "No headers present",
			headers:        map[string][]byte{},
			expectedResult: nil,
		},
		{
			name: "Empty X-Forwarded-For and X-Real-IP",
			headers: map[string][]byte{
				utils.HeaderXForwardedFor: []byte(""),
				utils.HeaderXRealIP:       []byte(""),
			},
			expectedResult: nil,
		},
		{
			name: "Only X-Real-IP with empty X-Forwarded-For",
			headers: map[string][]byte{
				utils.HeaderXForwardedFor: []byte(""),
				utils.HeaderXRealIP:       []byte("172.16.0.1"),
			},
			expectedResult: []byte("172.16.0.1"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockPeekable{headers: tc.headers}
			result := utils.RealIP(mock)
			req.Equal(tc.expectedResult, result)
		})
	}
}
