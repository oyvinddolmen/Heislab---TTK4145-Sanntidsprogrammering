package localip

import (
	"net"
	"strings"
)

func LocalIP() (string, error) {
	connection, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
	if err != nil {
		return "", err
	}
	defer connection.Close()

	localIP := strings.Split(connection.LocalAddr().String(), ":")[0]
	return localIP, nil
}
