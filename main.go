package main

import (
	"net"
)

var buf [1280]byte
func main() {
    // Listen on port 8080
    listenAddress, err := net.ResolveUDPAddr("udp", ":8080")
    if err != nil {
        panic(err)
    }

    conn, err := net.ListenUDP("udp", listenAddress)
    if err != nil {
        panic(err)
    }
    defer conn.Close()

	targetAddress, err := net.ResolveUDPAddr("udp", ":2456")
    if err != nil {
        panic(err)
    }
    // Set up a map to store the addresses of clients
    var clientAddr *net.UDPAddr

    
    for {
		// Read incoming packets and print them
        n, srcAddr, err := conn.ReadFromUDP(buf[:])
        if err != nil {
            panic(err)
        }
		clientAddr = srcAddr

		// fmt.Printf("%d\n", n)

		// Forward packets to proxy target
		_, err = conn.WriteToUDP(buf[:n], targetAddress)
		if err != nil {
			panic(err)
		}

		n, err = conn.Read(buf[:])
        if err != nil {
            panic(err)
        }

		// fmt.Printf("%d\n", n)

        // Write the response to the client
        _, err = conn.WriteToUDP(buf[:n], clientAddr)
        if err != nil {
            panic(err)
        }
    }
}
