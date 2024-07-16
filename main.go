package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/fuglesteg/timid/docker"
	"github.com/fuglesteg/timid/envInit"
	"github.com/fuglesteg/timid/proxy"
	"github.com/fuglesteg/timid/verboseLog"
)

// TODO: Experiment with proxy buffering packets until container starts
// TODO: Support multiple ports
var proxyServer *proxy.Proxy
var dockerController *docker.DockerController
var containerProcedureRunning = false
var oneMinuteDuration, _ = time.ParseDuration("1m")
var containerGroup *docker.ContainerGroup = new(docker.ContainerGroup)

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

	verbosityKey      = envInit.EnvKey("TIMID_LOG_VERBOSITY")
	containerNameKey  = envInit.EnvKey("TIMID_CONTAINER_NAME")
	containerGroupKey = envInit.EnvKey("TIMID_CONTAINER_GROUP")
)


func main() {
	verboseLog.Vlogf(1, "Starting...")
	initDockerController()
	initEnvVariables()

	verboseLog.Vlogf(1, "Proxy port = %d, Target address = %s\n",
		proxyPort, targetAddress)

	var err error
	proxyServer, err = proxy.NewProxy(proxyPort, targetAddress, connectionTimeoutDelay)

	verboseLog.Checkreport(1, err)

	if dockerController != nil {
		go func() {
			for {
				L: for {
					select {
					case <- proxyServer.OnConnection:
						go startContainers()
						break L
					case <- time.After(5 * time.Second):
						proxyServer.CleanUnusedConnections()
						shutdownContainerIfNoConnections(proxyServer)
						break L
					}
				}
			}
		}()
	}
	proxyServer.RunProxy()
}

func initDockerController() {
	containerName, containerErr := containerNameKey.GetEnvString()
	containerGroupName, groupErr := containerGroupKey.GetEnvString()
	if containerErr != nil && groupErr != nil {
		verboseLog.Checkreport(1, errors.New("Docker functionality disabled:"))
		groupErr = fmt.Errorf("Docker group functionality could not be initialized: %s:", groupErr)
		containerErr = fmt.Errorf("Docker container functionality could not be initialized %s:", containerErr)
		verboseLog.Checkreport(1, groupErr)
		verboseLog.Checkreport(1, containerErr)
		return
	}
	dockerController = docker.NewDockerController()
	var err error
	if containerName != "" {
		container, err := dockerController.NewContainer(containerName)
		containerGroup = docker.NewContainerGroup(containerName, []*docker.Container{container}, dockerController)
		if err != nil {
			panic(fmt.Errorf("Failed to initialize Docker functionality: %s", err))
		}
	} else {
		containerGroup, err = dockerController.NewContainerGroup(containerGroupName)
		if err != nil {
			panic(fmt.Errorf("Failed to initialize Docker functionality: %s", err))
		}
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
			verboseLog.Checkreport(4, fmt.Errorf("Containers will never be stopped %w", err))
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
	fmt.Printf("Logging verbosity: %d \n", verboseLog.Verbosity)
	connectionTimeoutDelay, err = connectionTimeoutDelayKey.GetEnvDurationOrFallback(time.Duration(5 * time.Second))
	if err != nil {
		verboseLog.Checkreport(4, fmt.Errorf("Proxy connection timeout delay not set: %w", err))
	}
}

func startContainers() {
	if containerGroup.AnyContainerIsPaused() {
		verboseLog.Vlogf(1, "Unpausing containers")
		containerGroup.Unpause()
	}
	if containerGroup.AnyContainerIsStopped() {
		verboseLog.Vlogf(1, "Starting containers")
		containerGroup.Start()
	}
}

func shutdownContainerIfNoConnections(proxy *proxy.Proxy) {
	if containerProcedureRunning ||
		proxy.GetConnectionsAmount() > 0 ||
		containerGroup.AllContainersAreStopped() || 
		containerGroup.AllContainersArePaused() {
		return
	}

	if pauseContainer {
		pauseContainerProcedure(containerShutdownDelay)
	} else {
		shutdownContainerProcedure(containerShutdownDelay)
	}
}

func pauseContainerProcedure(delay time.Duration) {
	verboseLog.Vlogf(1, "Pausing containers in group %s, after delay of %s",
		containerGroup.Name,
		containerShutdownDelay.String())
	containerProcedure(func() {
		containerGroup.Pause()
		verboseLog.Vlogf(1, "Containers in group %s paused", containerGroup.Name)
		if pauseDuration != 0 {
			containerProcedureRunning = false
			shutdownContainerProcedure(pauseDuration)
		}
		return
	}, delay)
}

func shutdownContainerProcedure(delay time.Duration) {
	verboseLog.Vlogf(1, "Stopping containers in group %s, after delay of %s",
		containerGroup.Name,
		delay.String())
	containerProcedure(func() {
		containerGroup.Stop()
		verboseLog.Vlogf(1, "Containers in group %s stopped", containerGroup.Name)
		return
	}, delay)
}

func containerProcedure(procedure func(), delay time.Duration) {
	if containerProcedureRunning {
		return
	}
	go func() {
		containerProcedureRunning = true
		defer func() { containerProcedureRunning = false }()
		for {
			select {
			case <-proxyServer.OnConnection: {
					verboseLog.Vlogf(1, "Connection detected, aborting pause/shutdown procedure")
					return
				}
			case <-time.After(delay): {
					procedure()
					return
				}
			}
		}
	}()
}
