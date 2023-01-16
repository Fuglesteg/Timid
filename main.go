package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/fuglesteg/valheim-server-sleeper/docker"
	"github.com/fuglesteg/valheim-server-sleeper/proxy"
	"github.com/fuglesteg/valheim-server-sleeper/verboseLog"
)

/* TODO: Control docker container, starts when player connects,
/* shut down when no users are connected*/

var proxyServer *proxy.Proxy

func main() {
	var iverb *int = flag.Int("v", 1, "Verbosity (0-6)")
	flag.Parse()
	verboseLog.Verbosity = *iverb

    proxyPort := 8080
    targetAddress := "localhost:2456"

	verboseLog.Vlogf(3, "Proxy port = %d, Server address = %s\n",
		proxyPort, targetAddress)

    timeoutDelay, _ := time.ParseDuration("2m");
    containerShutdownDelay, _ := time.ParseDuration("1m")
    proxyServer, err := proxy.NewProxy(proxyPort, targetAddress, timeoutDelay)
    if err != nil {
        verboseLog.Checkreport(5, err)
    }
    proxyServer.Start()
    dockerController := docker.NewDockerController()
    container := dockerController.NewContainer("valheim-valheim-1")
    containerProcedureRunning := false

    shutdownContainerIfNoConnections := func() {
        connectionsAmount := proxyServer.GetConnectionsAmount()
        noConnections := connectionsAmount <= 0
        if !noConnections {
            return
        }
        if containerProcedureRunning {
            return
        }
        if !dockerController.ContainerIsRunning(container) {
            return 
        }
        go func() {
            fmt.Println("Starting container shutdown timer")
            containerProcedureRunning = true
            defer func() {containerProcedureRunning = false}()
            time.Sleep(containerShutdownDelay)
            if proxyServer.GetConnectionsAmount() > 0 {
                return
            }
            fmt.Println("stopping container")
            dockerController.StopContainer(container)
        }()
    }

    startContainerIfConnections := func() {
        connectionsExist := proxyServer.GetConnectionsAmount() > 0
        if !connectionsExist {
            return
        }
        if containerProcedureRunning {
            return
        }
        if dockerController.ContainerIsRunning(container) {
            return 
        }
        dockerController.StartContainer(container)
    }

    for {
        time.Sleep(1 * time.Second)
        proxyServer.CleanUnusedConnections()
        fmt.Println(proxyServer.GetConnectionsAmount())
        shutdownContainerIfNoConnections();
        startContainerIfConnections();
    }
}
