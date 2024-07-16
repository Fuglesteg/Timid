package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fuglesteg/timid/docker"
	"github.com/fuglesteg/timid/proxy"
	"github.com/fuglesteg/timid/verboseLog"
)

type Api struct {
	ProxyServer *proxy.Proxy
	ContainerGroup *docker.ContainerGroup
	DockerController *docker.DockerController
}

type ContainerState string

const (
	Stopped ContainerState = "Stopped"
	Paused = "Paused"
	Running = "Running"
)

type Container struct {
	Id string `json:"id"`
	Name string `json:"name"`
	State ContainerState `json:"state"`
}

type ContainerGroup struct {
	Name string `json:"name"`
	State ContainerState `json:"state"`
}

type Info struct {
	Connections int `json:"connections"`
	ContainerGroup ContainerGroup `json:"containerGroup"`
}

type Proxy struct {
	Connections int `json:"connections"`
	Port int `json:"port"`
	TargetAddress string `json:"targetAddress"`
}

func (api Api) getContainerGroupState() ContainerState {
	if api.ContainerGroup.AllContainersArePaused() {
		return Paused
	} else if api.ContainerGroup.AllContainersAreStopped() {
		return Stopped
	} else {
		return Running
	}
}

func (api Api) getContainerState(containerId string) ContainerState {
	if api.DockerController.ContainerIsPaused(containerId) {
		return Paused
	} else if api.DockerController.ContainerIsRunning(containerId) {
		return Running
	} else {
		return Stopped
	}
}

func (api Api) mapContainerToContainerDTO(container docker.Container) Container {
	return Container {
		Name: container.Name,
		Id: container.ID,
		State: api.getContainerState(container.ID),
	}
}

func writeJsonToResponse(w http.ResponseWriter, value any) {
	bytes, err := json.Marshal(value);
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		verboseLog.Checkreport(2, err)
		return
	}

	w.Write(bytes)
}

func (api Api) Init(port int) {
	verboseLog.Vlogf(2, "Starting REST API")
	mux := http.NewServeMux()

	mux.HandleFunc("GET /info", func(w http.ResponseWriter, r *http.Request) {
		info := Info {
			Connections: api.ProxyServer.GetConnectionsAmount(),
			ContainerGroup: ContainerGroup {
				Name: api.ContainerGroup.Name,
				State: api.getContainerGroupState(),
			},
		}

		writeJsonToResponse(w, info);
	})

	mux.HandleFunc("POST /proxy/trigger", func(w http.ResponseWriter, r *http.Request) {
		api.ProxyServer.OnConnection <- 1
	})

	mux.HandleFunc("GET /proxy", func(w http.ResponseWriter, r *http.Request) {
		proxy := Proxy {
			Connections: api.ProxyServer.GetConnectionsAmount(),
			Port: api.ProxyServer.GetPort(),
			TargetAddress: api.ProxyServer.GetTargetAddress(),
		}

		writeJsonToResponse(w, proxy)
	})

	mux.HandleFunc("GET /containers", func(w http.ResponseWriter, r *http.Request) {
		containers := api.ContainerGroup.GetContainers()
		var containerDTOs []*Container
		for _, container := range containers {
			containerDTO := api.mapContainerToContainerDTO(*container)
			containerDTOs = append(containerDTOs, &containerDTO)
		}

		writeJsonToResponse(w, containerDTOs);
	})

	mux.HandleFunc("GET /containers/{containerId}", func(w http.ResponseWriter, r *http.Request) {
		containerId := r.PathValue("containerId")
		var containerDTO Container
		containers := api.ContainerGroup.GetContainers()
		for _, container := range containers {
			if containerId == container.ID {
				containerDTO = api.mapContainerToContainerDTO(*container)
			}
		}

		writeJsonToResponse(w, containerDTO);
	})

	mux.HandleFunc("POST /containers/start", func(w http.ResponseWriter, r *http.Request) {
		api.ContainerGroup.Start()
	})

	mux.HandleFunc("POST /containers/stop", func(w http.ResponseWriter, r *http.Request) {
		api.ContainerGroup.Stop()
	})

	mux.HandleFunc("POST /containers/pause", func(w http.ResponseWriter, r *http.Request) {
		api.ContainerGroup.Pause()
	})

	mux.HandleFunc("POST /containers/{containerId}/start", func(w http.ResponseWriter, r *http.Request) {
		containerId := r.PathValue("containerId")
		api.DockerController.StartContainer(containerId)
	})

	mux.HandleFunc("POST /containers/{containerId}/stop", func(w http.ResponseWriter, r *http.Request) {
		containerId := r.PathValue("containerId")
		api.DockerController.StopContainer(containerId)
	})

	mux.HandleFunc("POST /containers/{containerId}/pause", func(w http.ResponseWriter, r *http.Request) {
		containerId := r.PathValue("containerId")
		api.DockerController.PauseContainer(containerId)
	})

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
			panic(fmt.Errorf("Failed to initialize REST API: %s", err))
		}
	}()
}
