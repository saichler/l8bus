package common

import (
	"errors"
	"net"
	"time"
)

// Read data from socket
func Read(conn net.Conn) ([]byte, error) {
	// read 8 bytes, e.g. long, hinting of the size of the byte array
	sizebytes, err := ReadSize(8, conn)
	if sizebytes == nil || err != nil {
		return nil, err
	}
	// Translate the 8 byte array into int64
	size := Bytes2Long(sizebytes)
	// If the size is larger than the MAX Data Size, return an error
	// this is to protect against overflowing the buffers
	// When data to send is > the max data size, one needs to split the data into chunks at a higher level
	if size > NetConfig.MaxDataSize {
		return nil, errors.New("Max Size Exceeded!")
	}
	// Read the bunch of bytes according to the size from the socket
	data, err := ReadSize(int(size), conn)
	return data, err
}

func ReadSize(size int, conn net.Conn) ([]byte, error) {
	data := make([]byte, size)
	n, e := conn.Read(data)
	if e != nil {
		return nil, errors.New("Failed to read data size:" + e.Error())
	}

	if n < size {
		if n == 0 {
			time.Sleep(time.Second)
		}
		data = data[0:n]
		left, e := ReadSize(size-n, conn)
		if e != nil {
			return nil, errors.New("Failed to read packet size:" + e.Error())
		}
		data = append(data, left...)
	}
	return data, nil
}
