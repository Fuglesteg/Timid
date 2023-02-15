package main

import (
	"fmt"
	"time"

	"github.com/fuglesteg/timid/docker"
	"github.com/fuglesteg/timid/envInit"
	"github.com/fuglesteg/timid/proxy"
	"github.com/fuglesteg/timid/verboseLog"
)

// TODO: Implement option for containers to shut down if paused for long enough
// TODO: Experiment with proxy buffering packets until container starts
// TODO: Support multiple containers
// TODO: Support multiple ports

var proxyServer *proxy.Proxy
var dockerController *docker.DockerController
var container *docker.Container
var containerProcedureRunning = false
var oneMinuteDuration, _ = time.ParseDuration("1m")

var (
	pauseContainerKey = envInit.EnvKey("TIMID_PAUSE_CONTAINER")
	pauseContainer    bool

	pauseDurationKey = envInit.EnvKey("TIMID_PAUSE_DURATION")
	pauseDuration    time.Duration

	containerShutdownDelayKey = envInit.EnvKey("TIMID_CONTAINER_SHUTDOWN_DELAY")
	containerShutdownDelay    time.Duration

	targetAddressKey = envInit.EnvKey("TIMID_TARGET_ADDRESS")
	targetAddress    string

	proxyPortKey = envInit.EnvKey("TIMID_PORT")
	proxyPort    int

	connectionTimeoutDelayKey = envInit.EnvKey("TIMID_CONNECTION_TIMEOUT_DELAY")
	connectionTimeoutDelay    time.Duration

	verbosityKey     = envInit.EnvKey("TIMID_LOG_VERBOSITY")
	containerNameKey = envInit.EnvKey("TIMID_CONTAINER_NAME")
)

func main() {
	verboseLog.Vlogf(1, "Starting...")
	initDockerController()
	initEnvVariables()

	verboseLog.Vlogf(1, "Proxy port = %d, Target address = %s\n",
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
	containerName, err := containerNameKey.GetEnvString()
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

	containerShutdownDelay, err = containerShutdownDelayKey.GetEnvDurationOrFallback(oneMinuteDuration)
	if err != nil {
		verboseLog.Checkreport(4, fmt.Errorf("Container shutdown delay not set: %s", err))
	}

	pauseContainer, err = pauseContainerKey.GetEnvBoolOrFallback(false)
	if err != nil {
		verboseLog.Checkreport(4, fmt.Errorf("Docker controller will never pause a container: %s", err))
	}

	if pauseContainer {
		pauseDuration, err = pauseDurationKey.GetEnvDuration()
		if err != nil {
			verboseLog.Checkreport(4, fmt.Errorf("Container will never be stopped %w", err))
		}
	}
}

func initEnvVariables() {
	var err error
	proxyPort, err = proxyPortKey.GetEnvInt()
	targetAddress, err = targetAddressKey.GetEnvString()
	if err != nil {
		panic(fmt.Errorf("Failed to set proxy target address and/or listening port: %s", err))
	}
	verboseLog.Verbosity, err = verbosityKey.GetEnvIntOrFallback(1)
	if err != nil {
		verboseLog.Checkreport(4, fmt.Errorf("Logging verbosity not set: %w", err))
	}
	connectionTimeoutDelay, err = connectionTimeoutDelayKey.GetEnvDurationOrFallback(oneMinuteDuration)
	if err != nil {
		verboseLog.Checkreport(4, fmt.Errorf("Proxy connection timeout delay not set: %w", err))
	}
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

	if dockerController.ContainerIsPaused(container) {
		dockerController.UnpauseContainer(container)
	} else {
		dockerController.StartContainer(container)
	}
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
	if dockerController.ContainerIsPaused(container) {
		return
	}
	go func() {
		containerProcedureRunning = true
		defer func() { containerProcedureRunning = false }()
		verboseLog.Vlogf(1, "Shutting down container after delay of %s",
			containerShutdownDelay.String())
		time.Sleep(containerShutdownDelay)
		if proxy.GetConnectionsAmount() > 0 {
			return
		}
		if !dockerController.ContainerIsRunning(container) {
			return
		}
		if dockerController.ContainerIsPaused(container) {
			return
		}
		if pauseContainer {
			verboseLog.Vlogf(1, "Pausing container %s", container.Name)
			dockerController.PauseContainer(container)
			verboseLog.Vlogf(1, "Container paused")
			if pauseDuration != 0 {
				go func() {
					verboseLog.Vlogf(1, "Stopping container after delay of %s", pauseDuration.String())
					time.Sleep(pauseDuration)
					dockerController.StopContainer(container)
					verboseLog.Vlogf(1, "Container stopped")
				}()
			}
		} else {
			verboseLog.Vlogf(1, "Stopping container %s", container.Name)
			dockerController.StopContainer(container)
			verboseLog.Vlogf(1, "Container stopped")
		}
	}()
}
