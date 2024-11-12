package common

import (
	"errors"
	"net"
)

// Write data to socket
func Write(data []byte, conn net.Conn) error {
	// If the connection is nil, return an error
	if conn == nil {
		return errors.New("no Connection Available")
	}
	// Error is the data is too big
	if len(data) > int(NetConfig.MaxDataSize) {
		return errors.New("data is larger than MAX size allowed")
	}
	// Write the size of the data
	_, e := conn.Write(Long2Bytes(int64(len(data))))
	if e != nil {
		return e
	}
	// Write the actual data
	_, e = conn.Write(data)
	return e
}
