package chat

import (
	"io"
	"log"
	"strings"
	"sync"
)

const (
	crlf                 = "\x0d\x0a"
	telnetWILL      byte = 0xfb
	telnetDO        byte = 0xfd
	telnetDONT      byte = 0xfe
	telnetIAC       byte = 0xff
	telnetInterrupt byte = 0xf4
	telnetEraseLine byte = 0xf8
	telnetEraseChar byte = 0xf7
)

// Channel is a communication medium that sends and receives messages as strings
type Channel struct {
	sync.Mutex
	target     io.ReadWriteCloser
	readBuffer [64]byte
}

// NewChannel creates a new instance of Channel
func NewChannel(rw io.ReadWriteCloser) *Channel {
	return &Channel{
		target: rw,
	}
}

// ReadString from the target channel
func (c *Channel) ReadString() (string, error) {
	p := c.readBuffer[:]
	result := ""
	for {
		n, err := c.target.Read(p)
		if n == 0 && err == nil {
			return "", io.EOF
		}

		if n > 0 {
			result += string(p[:n])
		}
		if err != nil || n <= len(c.readBuffer) {
			break
		}
	}

	// check if we have a telnet command
	if len(result) > 0 && result[0] == telnetIAC {
		log.Printf("COMMAND: %q\n", result)
		if c.handleCommand([]byte(result)) == telnetInterrupt {
			return "", io.EOF
		}
	}

	return strings.TrimSuffix(result, crlf), nil
}

func (c *Channel) handleCommand(b []byte) byte {
	var command byte
	for i, v := range b {
		if v == telnetIAC {
			continue
		} else if v == telnetDO {
			writeBytes(c.target, telnetCommand(telnetWILL, b[i+1]))
		} else if v >= 240 && v <= 250 {
			command = v
		}
	}

	return command
}

func (c *Channel) Write(data string) {
	c.Lock()
	defer c.Unlock()

	// 1. clear the line
	writeBytes(c.target, []byte("\x1B[2K\x1B[1G"))
	// 2. send data
	writeBytes(c.target, []byte(data))
}

func writeBytes(target io.Writer, p []byte) {
	remaining := len(p)

	for remaining > 0 {
		n, err := target.Write(p)

		remaining -= n

		if err != nil {
			break
		}
	}
}

func telnetCommand(b ...byte) []byte {
	return append([]byte{telnetIAC}, b...)
}

// Close calls the underlining Close method for the target medium
func (c *Channel) Close() error {
	return c.target.Close()
}
