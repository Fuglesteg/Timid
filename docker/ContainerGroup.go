package docker

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
	group.forEachContainer(func(container *Container) {
		group.dockerController.StartContainer(container.ID)
	})
}

func (group *ContainerGroup) Stop() {
	group.forEachContainer(func(container *Container) {
		group.dockerController.StopContainer(container.ID)
	})
}

func (group *ContainerGroup) Pause() {
	group.forEachContainer(func(container *Container) {
		group.dockerController.PauseContainer(container.ID)
	})
}

func (group *ContainerGroup) Unpause() {
	group.forEachContainer(func(container *Container) {
		group.dockerController.UnpauseContainer(container.ID)
	})
}

func (group *ContainerGroup) AnyContainerIsPaused() bool {
	var isPaused bool = false
	group.forEachContainer(func(container *Container) {
		if group.dockerController.ContainerIsPaused(container.ID) {
			isPaused = true
		}
	})
	return isPaused
}

func (group *ContainerGroup) AnyContainerIsStopped() bool {
	var isStopped bool = false
	group.forEachContainer(func(container *Container) {
		if !group.dockerController.ContainerIsRunning(container.ID) {
			isStopped = true
		}
	})
	return isStopped
}

func (group *ContainerGroup) AnyContainerIsRunning() bool {
	var isRunning bool = false
	group.forEachContainer(func(container *Container) {
		if group.dockerController.ContainerIsRunning(container.ID) {
			isRunning = true
		}
	})
	return isRunning
}

func (group *ContainerGroup) AllContainersAreStopped() bool {
	isStopped := true
	group.forEachContainer(func(container *Container) {
		if group.dockerController.ContainerIsRunning(container.ID) {
			isStopped = false
		}
	})
	return isStopped
}

func (group *ContainerGroup) AllContainersArePaused() bool {
	isPaused := false
	group.forEachContainer(func(container *Container) {
		if group.dockerController.ContainerIsPaused(container.ID) {
			isPaused = true
		}
	})
	return isPaused
}

func (group *ContainerGroup) forEachContainer(function func(*Container)) {
	for _, container := range group.containers {
		function(container)
	}
}
