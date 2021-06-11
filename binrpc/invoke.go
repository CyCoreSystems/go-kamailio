package binrpc

import (
	"fmt"
	"io"
	"net"
)

type binRPCClientCodec struct {
	c io.ReadWriteCloser
}

func (c *binRPCClientCodec) ReadResponseBody(body interface{}) error {
	return nil
}

func (c *binRPCClientCodec) WriteRequest(name string) error {
	var methodName = BinRpcString(name)
	return methodName.Encode(c.c)
}

func newClientCodec(conn io.ReadWriteCloser) *binRPCClientCodec {
	return &binRPCClientCodec{
		c: conn,
	}
}

// InvokeMethod calls the given RPC method on the given host and port
func InvokeMethod(method string, host string, port string) error {

	conn, err := net.Dial("udp", host+":"+port)
	defer conn.Close() // nolint

	if err != nil {
		return fmt.Errorf("failed to connect to kamailio RPC server: %w", err)
	}

	codec := newClientCodec(conn)
	err = codec.WriteRequest(method)

	if err != nil {
		return fmt.Errorf("failed to invoke RPC method: %w", err)
	}

	return nil
}
