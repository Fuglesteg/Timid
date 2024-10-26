package docker

import "errors"

type ContainerGroup struct  {
	Name string
	containers []*Container
	dockerController *DockerController
}

func (group *ContainerGroup) GetContainers() []*Container {
	return group.containers
}

func NewContainerGroup(name string, containers []*Container, controller *DockerController) *ContainerGroup {
	return &ContainerGroup{Name: name, containers: containers, dockerController: controller}
}

func (group *ContainerGroup) Start() {
	for _, container := range group.containers {
		group.dockerController.StartContainer(container.ID)
	}
}

func (group *ContainerGroup) ContainerExists(containerId string) bool {
	exists := false
	for _, container := range group.containers {
		if container.ID == containerId {
			exists = true
			break
		}
	}
	return exists
}

func (group *ContainerGroup) ContainerIsPaused(containerId string) (bool, error) {
	if group.ContainerExists(containerId) {
		return group.dockerController.ContainerIsPaused(containerId), nil
	} else {
		return false, errors.New("Container does not exist in group")
	}
}

func (group *ContainerGroup) ContainerIsRunning(containerId string) (bool, error) {
	if group.ContainerExists(containerId) {
		return group.dockerController.ContainerIsRunning(containerId), nil
	} else {
		return false, errors.New("Container does not exist in group")
	}
}

func (group *ContainerGroup) StartContainer(containerId string) {
	if group.ContainerExists(containerId) {
		group.dockerController.StartContainer(containerId)
	}
}

func (group *ContainerGroup) Stop() {
	for _, container := range group.containers {
		group.dockerController.StopContainer(container.ID)
	}
}

func (group *ContainerGroup) StopContainer(containerId string) {
	if group.ContainerExists(containerId) {
		group.dockerController.StopContainer(containerId)
	}
}

func (group *ContainerGroup) Pause() {
	for _, container := range group.containers {
		group.dockerController.PauseContainer(container.ID)
	}
}

func (group *ContainerGroup) PauseContainer(containerId string) {
	if group.ContainerExists(containerId) {
		group.dockerController.PauseContainer(containerId)
	}
}

func (group *ContainerGroup) Unpause() {
	for _, container := range group.containers {
		group.dockerController.UnpauseContainer(container.ID)
	}
}

func (group *ContainerGroup) UnpauseContainer(containerId string) {
	if group.ContainerExists(containerId) {
		group.dockerController.UnpauseContainer(containerId)
	}
}

func (group *ContainerGroup) Restart() {
	for _, container := range group.containers {
		group.dockerController.RestartContainer(container.ID)
	}
}

func (group *ContainerGroup) RestartContainer(containerId string) {
	if group.ContainerExists(containerId) {
		group.dockerController.RestartContainer(containerId)
	}
}

func (group *ContainerGroup) AnyContainerIsPaused() bool {
	var isPaused bool = false
	for _, container := range group.containers {
		if group.dockerController.ContainerIsPaused(container.ID) {
			isPaused = true
		}
	}
	return isPaused
}

func (group *ContainerGroup) AnyContainerIsStopped() bool {
	var isStopped bool = false
	for _, container := range group.containers {
		if !group.dockerController.ContainerIsRunning(container.ID) {
			isStopped = true
		}
	}
	return isStopped
}

func (group *ContainerGroup) AnyContainerIsRunning() bool {
	var isRunning bool = false
	for _, container := range group.containers {
		if group.dockerController.ContainerIsRunning(container.ID) {
			isRunning = true
		}
	}
	return isRunning
}

func (group *ContainerGroup) AllContainersAreStopped() bool {
	isStopped := true
	for _, container := range group.containers {
		if group.dockerController.ContainerIsRunning(container.ID) {
			isStopped = false
		}
	}
	return isStopped
}

func (group *ContainerGroup) AllContainersArePaused() bool {
	isPaused := false
	for _, container := range group.containers {
		if group.dockerController.ContainerIsPaused(container.ID) {
			isPaused = true
		}
	}
	return isPaused
}

func (group *ContainerGroup) AllContainersAreRunning() bool {
	isRunning := true
	for _, container := range group.containers {
		if !group.dockerController.ContainerIsRunning(container.ID) {
			isRunning = false
		}
	}
	return isRunning
}
