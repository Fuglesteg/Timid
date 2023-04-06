package docker

type ContainerGroup struct  {
	Name string
	containers []*Container
	dockerController *DockerController
}

func NewContainerGroup(name string, containers []*Container, controller *DockerController) *ContainerGroup {
	return &ContainerGroup{Name: name, containers: containers, dockerController: controller}
}

func (group *ContainerGroup) Start() {
	group.forEachContainer(group.dockerController.StartContainer)
}

func (group *ContainerGroup) Stop() {
	group.forEachContainer(group.dockerController.StopContainer)
}

func (group *ContainerGroup) Pause() {
	group.forEachContainer(group.dockerController.PauseContainer)
}

func (group *ContainerGroup) Unpause() {
	group.forEachContainer(group.dockerController.UnpauseContainer)
}

func (group *ContainerGroup) AnyContainerIsPaused() bool {
	var isPaused bool = false
	group.forEachContainer(func(container *Container) {
		if group.dockerController.ContainerIsPaused(container) {
			isPaused = true
		}
	})
	return isPaused
}

func (group *ContainerGroup) AnyContainerIsStopped() bool {
	var isStopped bool = false
	group.forEachContainer(func(container *Container) {
		if !group.dockerController.ContainerIsRunning(container) {
			isStopped = true
		}
	})
	return isStopped
}

func (group *ContainerGroup) AnyContainerIsRunning() bool {
	var isRunning bool = false
	group.forEachContainer(func(container *Container) {
		if group.dockerController.ContainerIsRunning(container) {
			isRunning = true
		}
	})
	return isRunning
}

func (group *ContainerGroup) AllContainersAreStopped() bool {
	isStopped := true
	group.forEachContainer(func(container *Container) {
		if group.dockerController.ContainerIsRunning(container) {
			isStopped = false
		}
	})
	return isStopped
}

func (group *ContainerGroup) AllContainersArePaused() bool {
	isPaused := false
	group.forEachContainer(func(container *Container) {
		if group.dockerController.ContainerIsPaused(container) {
			isPaused = true
		}
	})
	return isPaused
}

func (group *ContainerGroup)forEachContainer(function func(*Container)) {
	for _, container := range group.containers {
		function(container)
	}
}
