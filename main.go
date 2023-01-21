package main

import (
	"os"
	"strconv"
	"time"

	"github.com/fuglesteg/valheim-server-sleeper/docker"
	"github.com/fuglesteg/valheim-server-sleeper/proxy"
	"github.com/fuglesteg/valheim-server-sleeper/verboseLog"
)

// TODO: Add ability to choose if container should be paused or shutdown
// NOTE: Should there be an option to stop container if paused for long enough?
// TODO: Experiment with proxy buffering packets until container starts

var proxyServer *proxy.Proxy
var dockerController *docker.DockerController = docker.NewDockerController()
var container = dockerController.NewContainer("valheim-valheim-1")
var containerProcedureRunning = false
var containerShutdownDelay, _ = time.ParseDuration("1m")
var targetAddress = "localhost:2456"
var proxyPort = 8080

func main() {
        initEnvVariables()

	verboseLog.Vlogf(3, "Proxy port = %d, Target address = %s\n",
		proxyPort, targetAddress)

	timeoutDelay, _ := time.ParseDuration("1m")
	proxyServer, err := proxy.NewProxy(proxyPort, targetAddress, timeoutDelay)
	if err != nil {
		verboseLog.Checkreport(5, err)
	}
	proxyServer.Start()

	for {
		time.Sleep(1 * time.Second)
		proxyServer.CleanUnusedConnections()
		shutdownContainerIfNoConnections()
		startContainerIfConnections()
	}
}

func initEnvVariables() {
        proxyPortEnv, err := strconv.Atoi(os.Getenv("PROXY_PORT"))
        if err == nil {
            proxyPort = proxyPortEnv
        }
        targetAddressEnv := os.Getenv("PROXY_TARGET_ADDRESS")
        if targetAddressEnv != "" {
            targetAddress = targetAddressEnv
        }
        containerShutdownDelayEnv, err := time.ParseDuration(os.Getenv("PROXY_CONTAINER_SHUTDOWN_DELAY"))
        if err == nil {
            containerShutdownDelay = containerShutdownDelayEnv
        }
        verbosity, err := strconv.Atoi(os.Getenv("PROXY_LOG_VERBOSITY"))
        if err == nil {
            verboseLog.Verbosity = verbosity
        }
}

func startContainerIfConnections() {
	connectionsExist := proxyServer.GetConnectionsAmount() > -1
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

func shutdownContainerIfNoConnections() {
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
		verboseLog.Vlogf(1, "Starting container shutdown timer")
		containerProcedureRunning = true
		defer func() { containerProcedureRunning = false }()
		verboseLog.Vlogf(1, "Shutting down container after delay of " +
			containerShutdownDelay.String())
		time.Sleep(containerShutdownDelay)
		if proxyServer.GetConnectionsAmount() > 0 {
			return
		}
		verboseLog.Vlogf(1, "stopping container")
		dockerController.StopContainer(container)
	}()
}
