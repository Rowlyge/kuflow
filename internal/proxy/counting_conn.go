package proxy

import "net"

// countingConn считает количество переданных байт.
type countingConn struct {
	net.Conn

	readBytes  uint64
	writeBytes uint64
}

func (c *countingConn) Read(
	p []byte,
) (int, error) {

	n, err := c.Conn.Read(p)

	c.readBytes += uint64(n)

	return n, err
}

func (c *countingConn) Write(
	p []byte,
) (int, error) {

	n, err := c.Conn.Write(p)

	c.writeBytes += uint64(n)

	return n, err
}
