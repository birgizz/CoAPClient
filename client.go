package coap

import (
	"net"
	"time"
)
const maxPktLen = 2000

const (
	// ResponseTimeout is the amount of time to wait for a
	// response.
	ResponseTimeout = time.Second * 2
	// ResponseRandomFactor is a multiplier for response backoff.
	ResponseRandomFactor = 1.5
	// MaxRetransmit is the maximum number of times a message will
	// be retransmitted.
	MaxRetransmit = 4
)

// Conn is a CoAP client connection.
type Conn struct {
	conn *net.UDPConn
	buf  []byte
}

// Dial connects a CoAP client.
func Dial(n, addr string) (*Conn, error) {
	uaddr, err := net.ResolveUDPAddr(n, addr)
	if err != nil {
		return nil, err
	}

	s, err := net.DialUDP("udp", nil, uaddr)
	if err != nil {
		return nil, err
	}

	return &Conn{s, make([]byte, maxPktLen)}, nil
}

// Send a message.  Get a response if there is one.
func (c *Conn) Send(req Message) (*Message, error) {
	err := Transmit(c.conn, nil, req)
	if err != nil {
		return nil, err
	}

	if !req.IsConfirmable() {
		return nil, nil
	}

	rv, err := Receive(c.conn, c.buf)
	if err != nil {
		return nil, err
	}

	return &rv, nil
}

// Receive a message.
func (c *Conn) Receive() (*Message, error) {
	rv, err := Receive(c.conn, c.buf)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}


func Transmit(l *net.UDPConn, a *net.UDPAddr, m Message) error {
	d, err := m.MarshalBinary()
	if err != nil {
		return err
	}

	if a == nil {
		_, err = l.Write(d)
	} else {
		_, err = l.WriteTo(d, a)
	}
	return err
}

// Receive a message.
func Receive(l *net.UDPConn, buf []byte) (Message, error) {
	l.SetReadDeadline(time.Now().Add(ResponseTimeout))

	nr, _, err := l.ReadFromUDP(buf)
	if err != nil {
		return Message{}, err
	}
	return ParseMessage(buf[:nr])
}

