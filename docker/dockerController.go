package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/fuglesteg/timid/verboseLog"
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

func (controller *DockerController) StopContainer(container *Container) {
	err := controller.client.ContainerKill(context.Background(), container.ID, "SIGTERM")
	if err != nil {
		verboseLog.Checkreport(1, err)
	}
}

func (controller *DockerController) PauseContainer(container *Container) {
	err := controller.client.ContainerPause(context.Background(), container.ID)
	if err != nil {
		verboseLog.Checkreport(1, err)
	}
}

func (controller *DockerController) UnpauseContainer(container *Container) {
	err := controller.client.ContainerUnpause(context.Background(), container.ID)
	if err != nil {
		verboseLog.Checkreport(1, err)
	}
}

func (controller *DockerController) StartContainer(container *Container) {
	err := controller.client.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
	if err != nil {
		verboseLog.Checkreport(1, err)
	}
}

func (controller *DockerController) ContainerIsRunning(container *Container) bool {
	info, err := controller.client.ContainerInspect(context.Background(), container.ID)
	if err != nil {
		verboseLog.Checkreport(1, err)
	}
	return info.State.Running
}

func (controller *DockerController) ContainerIsPaused(container *Container) bool {
	info, err := controller.client.ContainerInspect(context.Background(), container.ID)
	if err != nil {
		verboseLog.Checkreport(1, err)
	}
	return info.State.Paused
}

func (controller *DockerController) NewContainer(containerName string) (*Container, error) {
	filterArgs := filters.NewArgs(
		filters.Arg("name", containerName),
	)
	listOptions := types.ContainerListOptions{All: true, Filters: filterArgs}
	containers, err :=
		controller.client.ContainerList(context.Background(), listOptions)
	if err != nil {
		return nil, err
	}
	container := &Container{Name: containerName, ID: containers[0].ID}
	return container, nil
}
