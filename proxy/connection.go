package proxy

import (
	"net"
	"time"

	"github.com/fuglesteg/timid/verboseLog"
)

// Information maintained for each client/server connection
type connection struct {
	ClientAddr *net.UDPAddr // Address of the client
	ServerConn *net.UDPConn // UDP connection to server
        LastUsed   *time.Time
}

func (connection *connection) UpdateLastUsed() {
    timeNow := time.Now()
    connection.LastUsed = &timeNow
}

// Generate a new connection by opening a UDP connection to the server
func newConnection(srvAddr, cliAddr *net.UDPAddr) *connection {
	conn := new(connection)
	conn.ClientAddr = cliAddr
	srvUdp, err := net.DialUDP("udp", nil, srvAddr)
	if verboseLog.Checkreport(1, err) {
		return nil
	}
	conn.ServerConn = srvUdp
	return conn
}
