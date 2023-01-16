package proxy

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/fuglesteg/valheim-server-sleeper/verboseLog"
)

type Proxy struct {
    // Connection used by clients as the proxy server
    proxyConn *net.UDPConn

    // Address of server
    serverAddr *net.UDPAddr

    // Mapping from client addresses (as host:port) to connection
    clientDict map[string]*connection

    // Mutex used to serialize access to the dictionary
    dmutex *sync.Mutex

    timeOutDelay time.Duration
}

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

func (proxy *Proxy)CleanUnusedConnections() {
    for _, connection := range proxy.clientDict {
        timeoutReached := time.Since(*connection.LastUsed).Seconds() > proxy.timeOutDelay.Seconds() // Gives wrong time if not using seconds
        if (timeoutReached) {
            delete(proxy.clientDict, connection.ClientAddr.String())
        }
    }
}

func (proxy *Proxy)GetConnectionsAmount() int {
    return len(proxy.clientDict)
}

// Generate a new connection by opening a UDP connection to the server
func newConnection(srvAddr, cliAddr *net.UDPAddr) *connection {
	conn := new(connection)
	conn.ClientAddr = cliAddr
	srvudp, err := net.DialUDP("udp", nil, srvAddr)
	if verboseLog.Checkreport(1, err) {
		return nil
	}
	conn.ServerConn = srvudp
	return conn
}

func NewProxy(proxyPort int, targetAddress string, connectionTimeoutDelay time.Duration) (proxy *Proxy, err error) {
    proxy = new(Proxy)
    proxy.clientDict = make(map[string]*connection)
    proxy.dmutex = new(sync.Mutex)
    proxy.timeOutDelay = connectionTimeoutDelay
    err = proxy.setup(proxyPort, targetAddress)

    return proxy, err
}

func (proxy *Proxy)Start() {
    go proxy.runProxy()
}

func (proxy *Proxy)setup(port int, hostport string) error {
	// Set up Proxy
	saddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if verboseLog.Checkreport(1, err) {
		return err
	}
	pudp, err := net.ListenUDP("udp", saddr)
	if verboseLog.Checkreport(1, err) {
		return err
	}
	proxy.proxyConn = pudp
	verboseLog.Vlogf(2, "Proxy serving on port %d\n", port)

	// Get server address
	srvaddr, err := net.ResolveUDPAddr("udp", hostport)
	if verboseLog.Checkreport(1, err) {
		return err
	}
	proxy.serverAddr = srvaddr
	verboseLog.Vlogf(2, "Connected to server at %s\n", hostport)
	return nil
}

func (proxy *Proxy)dlock() {
	proxy.dmutex.Lock()
}

func (proxy *Proxy)dunlock() {
	proxy.dmutex.Unlock()
}

// Go routine which manages connection from server to single client
func (proxy *Proxy)runConnection(conn *connection) {
	var buffer [1500]byte
	for {
		// Read from server
		n, err := conn.ServerConn.Read(buffer[0:])
		if verboseLog.Checkreport(1, err) {
			continue
		}
		// Relay it to client
		_, err = proxy.proxyConn.WriteToUDP(buffer[0:n], conn.ClientAddr)
		if verboseLog.Checkreport(1, err) {
			continue
		}
        conn.UpdateLastUsed()
		verboseLog.Vlogf(3, "Relayed '%s' from server to %s.\n",
			string(buffer[0:n]), conn.ClientAddr.String())
	}
}

// Routine to handle inputs to Proxy port
func (proxy *Proxy)runProxy() {
	var buffer [1500]byte
	for {
		n, cliaddr, err := proxy.proxyConn.ReadFromUDP(buffer[0:])
		if verboseLog.Checkreport(1, err) {
			continue
		}
		verboseLog.Vlogf(3, "Read '%s' from client %s\n",
			string(buffer[0:n]), cliaddr.String())
		saddr := cliaddr.String()
		proxy.dlock()
		conn, found := proxy.clientDict[saddr]
		if !found {
			conn = newConnection(proxy.serverAddr, cliaddr)
			if conn == nil {
				proxy.dunlock()
				continue
			}
			proxy.clientDict[saddr] = conn
            conn.UpdateLastUsed()
			proxy.dunlock()
			verboseLog.Vlogf(2, "Created new connection for client %s\n", saddr)
			// Fire up routine to manage new connection
			go proxy.runConnection(conn)
		} else {
			verboseLog.Vlogf(5, "Found connection for client %s\n", saddr)
			proxy.dunlock()
		}
		// Relay to server
		_, err = conn.ServerConn.Write(buffer[0:n])
		if verboseLog.Checkreport(1, err) {
			continue
		}
	}
}
