package ipc

import (
	"bufio"
	"net"
	"strings"

	"github.com/pkg/errors"
)

func CreateSocket() (net.Listener, error) {
	return net.Listen("tcp", ":42069")
}

func SocketAccept(ln net.Listener) (net.Conn, error) {
	return ln.Accept()
}

func ConnectToSocket() (net.Conn, error) {
	return net.Dial("tcp", "127.0.0.1:42069")
}

func Send(conn net.Conn, msg string) error {
	if _, err := conn.Write([]byte(msg + "\n")); err != nil {
		return errors.Wrap(err, "error while sending message to socket")
	}
	return nil
}

func WaitForMessage(conn net.Conn, msg string) error {
	buf, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return errors.Wrap(err, "error while recieving message from socket")
	}
	buf = strings.TrimSuffix(buf, "\n")
	if buf == msg {
		return nil
	} else {
		return errors.Errorf("Got %q while waiting for: %q", buf, msg)
	}
}
