package docker

import (
	"context"
	"errors"

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
	err := controller.client.ContainerStop(context.Background(), container.ID, nil)
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
		return false
	}
	return info.State.Status == "running" && !info.State.Paused
}

func (controller *DockerController) ContainerIsPaused(container *Container) bool {
	info, err := controller.client.ContainerInspect(context.Background(), container.ID)
	if err != nil {
		verboseLog.Checkreport(1, err)
		return false
	}
	return info.State.Status == "paused" || info.State.Paused
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
	containerId := containers[0].ID
	container := &Container{Name: containerName, ID: containerId}
	return container, nil
}

func (controller *DockerController) NewContainerGroup(groupName string) (*ContainerGroup, error) {
	filterArgs := filters.NewArgs(
		filters.Arg("label", "timid.group." + groupName),
	)
	listOptions := types.ContainerListOptions{All: true, Filters: filterArgs}
	filteredContainers, err :=
		controller.client.ContainerList(context.Background(), listOptions)
	if err != nil {
		return nil, err
	}
	if len(filteredContainers) <= 0 {
		return nil, errors.New("No containers found with label: " + "timid.group." + groupName)
	}
	var initializedContainers []*Container
	for _, container := range filteredContainers {
		container := &Container{Name: container.Names[0], ID: container.ID}
		initializedContainers = append(initializedContainers, container)
	}
	containerGroup := ContainerGroup {
		Name: groupName, 
		dockerController: controller, 
		containers: initializedContainers,
	}
	return &containerGroup, nil
}
