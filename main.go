package main

import (
	"fmt"
	"time"

	"github.com/fuglesteg/timid/docker"
	"github.com/fuglesteg/timid/envInit"
	"github.com/fuglesteg/timid/proxy"
	"github.com/fuglesteg/timid/verboseLog"
)

// TODO: Add ability to choose if container should pause or shut down
// NOTE: Should there be an option to stop container if paused for long enough?
// TODO: Experiment with proxy buffering packets until container starts

var proxyServer *proxy.Proxy
var dockerController *docker.DockerController
var container *docker.Container
var containerProcedureRunning = false

var (
	pauseContainer            = false
	containerShutdownDelay, _ = time.ParseDuration("1m")
	targetAddress             string
	proxyPort                 int
	connectionTimeoutDelay, _ = time.ParseDuration("1m")
)

func main() {
	initDockerController()
	initEnvVariables()

	verboseLog.Vlogf(3, "Proxy port = %d, Target address = %s\n",
		proxyPort, targetAddress)

	proxyServer, err := proxy.NewProxy(proxyPort, targetAddress, connectionTimeoutDelay)
	if err != nil {
		panic(err)
	}

	if dockerController != nil {
		go func() {
			for {
				time.Sleep(1 * time.Second)
				proxyServer.CleanUnusedConnections()
				shutdownContainerIfNoConnections(proxyServer)
				startContainerIfConnections(proxyServer)
			}
		}()
	}

	proxyServer.RunProxy()
}

func initDockerController() {
	containerName, err := envInit.GetEnvString("PROXY_CONTAINER_NAME")
	if err != nil {
		err = fmt.Errorf("Docker functionality disabled: %s", err)
		verboseLog.Checkreport(1, err)
		return
	}
	dockerController = docker.NewDockerController()
	container, err = dockerController.NewContainer(containerName)
	if err != nil {
		panic(fmt.Errorf("Failed to initialize Docker functionality: %s", err))
	}
	err = envInit.SetEnvDuration("PROXY_CONTAINER_SHUTDOWN_DELAY", &containerShutdownDelay)
	verboseLog.Checkreport(4, fmt.Errorf("Container shutdown delay not set: %s", err))
	err = envInit.SetEnvBool("PROXY_PAUSE_CONTAINER", &pauseContainer)
	verboseLog.Checkreport(4, fmt.Errorf("Docker controller will never pause a container: %s", err))
}

func initEnvVariables() {
	err := envInit.SetEnvInt("PROXY_PORT", &proxyPort)
	err = envInit.SetEnvString("PROXY_TARGET_ADDRESS", &targetAddress)
	if err != nil {
		panic(fmt.Errorf("Failed to set proxy target address and/or listening port: %s", err))
	}
	err = envInit.SetEnvInt("PROXY_LOG_VERBOSITY", &verboseLog.Verbosity)
	verboseLog.Checkreport(4, fmt.Errorf("Logging verbosity not set: %s", err))
	err = envInit.SetEnvDuration("PROXY_CONNECTION_TIMEOUT_DELAY", &connectionTimeoutDelay)
	verboseLog.Checkreport(4, fmt.Errorf("Proxy connection timeout delay not set: %s", err))
}

func startContainerIfConnections(proxy *proxy.Proxy) {
	connectionsExist := proxy.GetConnectionsAmount() > 0
	if !connectionsExist {
		return
	}
	if containerProcedureRunning {
		return
	}
	if dockerController.ContainerIsRunning(container) {
		return
	}
	verboseLog.Vlogf(1, "Detected connection, starting container")
	dockerController.StartContainer(container)
}

func shutdownContainerIfNoConnections(proxy *proxy.Proxy) {
	connectionsAmount := proxy.GetConnectionsAmount()
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
		verboseLog.Vlogf(1, "Shutting down container after delay of %s",
			containerShutdownDelay.String())
		time.Sleep(containerShutdownDelay)
		if proxy.GetConnectionsAmount() > 0 {
			return
		}
		verboseLog.Vlogf(1, "Stopping container")
		dockerController.StopContainer(container)
	}()
}
