package docker

import (
	"context"
	"errors"

	dContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/fuglesteg/timid/verboseLog"
)

type DockerController struct {
	client *client.Client
}

func NewDockerController() *DockerController {
	context := context.Background()
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		verboseLog.Checkreport(1, err)
	}
	client.NegotiateAPIVersion(context)
	dockerController := new(DockerController)
	dockerController.client = client
	return dockerController
}

func (controller *DockerController) StopContainer(containerId string) {
	err := controller.client.ContainerStop(context.Background(), containerId, dContainer.StopOptions{})
	if err != nil {
		verboseLog.Checkreport(1, err)
	}
}

func (controller *DockerController) PauseContainer(containerId string) {
	err := controller.client.ContainerPause(context.Background(), containerId)
	if err != nil {
		verboseLog.Checkreport(1, err)
	}
}

func (controller *DockerController) UnpauseContainer(containerId string) {
	err := controller.client.ContainerUnpause(context.Background(), containerId)
	if err != nil {
		verboseLog.Checkreport(1, err)
	}
}

func (controller *DockerController) StartContainer(containerId string) {
	err := controller.client.ContainerStart(context.Background(), containerId, dContainer.StartOptions{})
	if err != nil {
		verboseLog.Checkreport(1, err)
	}
}

func (controller *DockerController) ContainerIsRunning(containerId string) bool {
	info, err := controller.client.ContainerInspect(context.Background(), containerId)
	if err != nil {
		verboseLog.Checkreport(1, err)
		return false
	}
	return info.State.Status == "running" && !info.State.Paused
}

func (controller *DockerController) ContainerIsPaused(containerId string) bool {
	info, err := controller.client.ContainerInspect(context.Background(), containerId)
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
	listOptions := dContainer.ListOptions{All: true, Filters: filterArgs}
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
	listOptions := dContainer.ListOptions{All: true, Filters: filterArgs}
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
