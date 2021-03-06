package connector

import (
	"bytes"
	"net"
	"sync"

	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"golang.org/x/crypto/ssh"
)

// SSHConnection encapsulates the connection to the device
type SSHConnection struct {
	host   string
	client *ssh.Client
	conn   net.Conn
	mu     sync.Mutex
	done   chan struct{}
}

// RunCommand runs a command against the device
func (c *SSHConnection) RunCommand(cmd string) (string, error) {
	log.Debugf("Running command on %s:%s\n", c.host, cmd)
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		return "", errors.Errorf("Running command on %s:%s: Not connected.", c.host, cmd)
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", errors.Wrapf(err, "Running command on %s:%s: Coud not open session.", c.host, cmd)
	}
	defer session.Close()

	var b = &bytes.Buffer{}
	session.Stdout = b

	err = session.Run(cmd)
	if err != nil {
		return "", errors.Wrapf(err, "Running command on %s:%s: Coud not run command.", c.host, cmd)
	}
	// log.Debugf("Output for %s:%s\n", c.host, string(b.Bytes()))
	return string(b.Bytes()), nil
}

func (c *SSHConnection) isConnected() bool {
	return c.conn != nil
}

func (c *SSHConnection) terminate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.conn.Close()

	c.client = nil
	c.conn = nil
}

func (c *SSHConnection) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		c.client.Close()
	}

	c.done <- struct{}{}
	c.conn = nil
	c.client = nil
}

// Host returns the hostname connected to
func (c *SSHConnection) Host() string {
	return c.host
}
