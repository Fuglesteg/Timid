package proxy

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/fuglesteg/timid/verboseLog"
)

type Proxy struct {
	// Server address as string
	targetAddr string

	// Port to listen to
	port int

	// Connection used by clients as the proxy server
	proxyConn *net.UDPConn

	// Address of server
	serverAddr *net.UDPAddr

	// Mapping from client addresses (as host:port) to connection
	clientDict map[string]*connection

	// Mutex used to serialize access to the dictionary
	dmutex *sync.Mutex

	// Time until the proxy treats a connection as unused
	timeOutDelay time.Duration

	// Channel which reacts to connections
	OnConnection chan int
}

func NewProxy(proxyPort int, targetAddress string, connectionTimeoutDelay time.Duration) (*Proxy, error) {
	proxy := new(Proxy)
	proxy.clientDict = make(map[string]*connection)
	proxy.dmutex = new(sync.Mutex)
	proxy.timeOutDelay = connectionTimeoutDelay
	proxy.targetAddr = targetAddress
	proxy.port = proxyPort
	proxy.OnConnection = make(chan int)
	err := proxy.setup()

	return proxy, err
}

func (proxy *Proxy) GetPort() int {
	return proxy.port;
}

func (proxy *Proxy) GetTargetAddress() string {
	return proxy.targetAddr;
}

func (proxy *Proxy) CleanUnusedConnections() {
	go func() {
		for _, connection := range proxy.clientDict {
			timeoutReached := time.Since(*connection.LastUsed) > proxy.timeOutDelay
			if timeoutReached {
				delete(proxy.clientDict, connection.ClientAddr.String())
				verboseLog.Vlogf(2, "Removed unused connection for client: %s",
					connection.ClientAddr.String())
			}
		}
	}()
}

func (proxy *Proxy) GetConnectionsAmount() int {
	return len(proxy.clientDict)
}

func (proxy *Proxy) Start() {
	go proxy.RunProxy()
}

func (proxy *Proxy) setup() error {
	proxy.dlock()
	defer proxy.dunlock()
	// Set up Proxy
	if proxy.proxyConn == nil {
		saddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", proxy.port))
		if verboseLog.Checkreport(1, err) {
			return err
		}
		pudp, err := net.ListenUDP("udp", saddr)
		if verboseLog.Checkreport(1, err) {
			return err
		}
		proxy.proxyConn = pudp
		verboseLog.Vlogf(1, "Proxy serving on port %d\n", proxy.port)
	}

	if proxy.serverAddr == nil {
		// Get server address
		srvaddr, err := net.ResolveUDPAddr("udp", proxy.targetAddr)
		if verboseLog.Checkreport(1, err) {
			proxy.serverAddr = nil
			return err
		}
		proxy.serverAddr = srvaddr
		verboseLog.Vlogf(1, "Connected to server at %s\n", proxy.targetAddr)
	}
	return nil
}

func (proxy *Proxy) dlock() {
	proxy.dmutex.Lock()
}

func (proxy *Proxy) dunlock() {
	proxy.dmutex.Unlock()
}

// Go routine which manages connection from server to single client
func (proxy *Proxy) runConnection(conn *connection) {
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
func (proxy *Proxy) RunProxy() {
	var buffer [1500]byte
	for {
		n, clientAddr, err := proxy.proxyConn.ReadFromUDP(buffer[0:])
		if verboseLog.Checkreport(1, err) {
			continue
		}
		proxy.OnConnection <-1
		verboseLog.Vlogf(3, "Read '%s' from client %s\n",
			string(buffer[0:n]), clientAddr.String())
		clientAddressString := clientAddr.String()
		proxy.dlock()
		if proxy.serverAddr == nil {
			proxy.dunlock()
			err := proxy.setup()
			if verboseLog.Checkreport(1, err) {
				continue
			}
			proxy.dlock()
		}
		conn, found := proxy.clientDict[clientAddressString]
		if !found {
			conn = newConnection(proxy.serverAddr, clientAddr)
			if conn == nil {
				proxy.dunlock()
				continue
			}
			proxy.clientDict[clientAddressString] = conn
			conn.UpdateLastUsed()
			proxy.dunlock()
			verboseLog.Vlogf(2, "Created new connection for client %s\n", clientAddressString)
			// Fire up routine to manage new connection
			go proxy.runConnection(conn)
		} else {
			verboseLog.Vlogf(5, "Found connection for client %s\n", clientAddressString)
			proxy.dunlock()
		}
		// Relay to server
		_, err = conn.ServerConn.Write(buffer[0:n])
		if verboseLog.Checkreport(1, err) {
			continue
		}
	}
}
