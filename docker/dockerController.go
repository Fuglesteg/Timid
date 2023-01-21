package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/fuglesteg/valheim-server-sleeper/verboseLog"
)

type DockerController struct {
	client *client.Client
}

func NewDockerController() *DockerController {
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		verboseLog.Checkreport(1, err)
	}
	dockerController := new(DockerController)
	dockerController.client = client
	return dockerController
}

func (controller *DockerController) StopContainer(container *container) {
    err := controller.client.ContainerKill(context.Background(), container.ID, "SIGTERM")   
    if err != nil {
        verboseLog.Checkreport(1, err)
    }
}

func (controller *DockerController) PauseContainer(container *container) {
    err := controller.client.ContainerPause(context.Background(), container.ID)
    if err != nil {
        verboseLog.Checkreport(1, err)
    }
}

func (controller *DockerController) UnpauseContainer(container *container) {
    err := controller.client.ContainerUnpause(context.Background(), container.ID) 
    if err != nil {
        verboseLog.Checkreport(1, err)
    }
}

func (controller *DockerController) StartContainer(container *container) {
    err := controller.client.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
    if err != nil {
        verboseLog.Checkreport(1, err)
    }
}

func (controller *DockerController) ContainerIsRunning(container *container) bool{
    info, err := controller.client.ContainerInspect(context.Background(), container.ID)    
    if err != nil {
        verboseLog.Checkreport(1, err)
    }
    return info.State.Running
}

func (controller *DockerController) NewContainer(containerName string) *container {
    filterArgs := filters.NewArgs(
        filters.Arg("name", containerName),
    )
    listOptions := types.ContainerListOptions{All: true, Filters: filterArgs}
    containers, err := 
        controller.client.ContainerList(context.Background(), listOptions)
    if err != nil {
        verboseLog.Checkreport(1, err)
    }
    container := &container{Name: containerName, ID: containers[0].ID}
    return container
}
