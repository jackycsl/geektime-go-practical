package net

import (
	"errors"
	"io"
	"net"
)

func Serve(network, addr string) error {
	listener, err := net.Listen(network, addr)
	// 比较常见的就是端口被占用
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			if er := handleConn(conn); er != nil {
				conn.Close()
			}
		}()
	}
}

func handleConn(conn net.Conn) error {
	for {
		bs := make([]byte, 8)
		_, err := conn.Read(bs)
		if err == net.ErrClosed || err == io.EOF || err == io.ErrUnexpectedEOF {
			return err
		}
		// 这种是可以挽救的
		if err != nil {
			continue
		}
		// if n != 8 {
		// 	return errors.New("micro: 没读够数据")
		// }

		res := handleMsg(bs)
		_, err = conn.Write(res)
		// 这种是可以挽救的
		if err != nil {
			continue
		}
		// if n != len(res) {
		// 	return errors.New("micro: 没写完数据")
		// }
	}
}

func handleConnV1(conn net.Conn) error {
	for {
		bs := make([]byte, 8)
		n, err := conn.Read(bs)
		if err != nil {
			return err
		}
		if n != 8 {
			return errors.New("micro: 没读够数据")
		}
		res := handleMsg(bs)
		n, err = conn.Write(res)
		// 这种是可以挽救的
		if err != nil {
			return err
		}
		if n != len(res) {
			return errors.New("micro: 没写完数据")
		}
	}
}

func handleMsg(req []byte) []byte {
	res := make([]byte, len(req))
	copy(res[:len(req)], req)
	copy(res[len(req):], req)
	return res
}
