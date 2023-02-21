package docker

type ContainerGroup struct  {
	Name string
	Containers []*Container
	DockerController *DockerController
}

func (group *ContainerGroup) Start() {
	group.forEachContainer(group.DockerController.StartContainer)
}

func (group *ContainerGroup) Stop() {
	group.forEachContainer(group.DockerController.StopContainer)
}

func (group *ContainerGroup) Pause() {
	group.forEachContainer(group.DockerController.PauseContainer)
}

func (group *ContainerGroup) Unpause() {
	group.forEachContainer(group.DockerController.UnpauseContainer)
}

func (group *ContainerGroup) AnyContainerIsPaused() bool {
	var isPause bool = false
	group.forEachContainer(func(container *Container) {
		if group.DockerController.ContainerIsPaused(container) {
			isPause = true
		}
	})
	return isPause
}

func (group *ContainerGroup) AnyContainerIsStopped() bool {
	var isStopped bool = false
	group.forEachContainer(func(container *Container) {
		if !group.DockerController.ContainerIsRunning(container) {
			isStopped = true
		}
	})
	return isStopped
}

func (group *ContainerGroup)forEachContainer(function func(*Container)) {
	for _, container := range group.Containers {
		function(container)
	}
}
