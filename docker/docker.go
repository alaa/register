package docker

import (
	"log"
	"sync"

	"github.com/fsouza/go-dockerclient"
)

type Docker struct {
	client *docker.Client
}

type Containers []*docker.Container

func New() *Docker {
	endpoint := "unix:///var/run/docker.sock"
	dockerClient, err := docker.NewClient(endpoint)
	if err != nil {
		log.Fatal(err)
	}

	return &Docker{
		client: dockerClient,
	}
}

func (d *Docker) ListRunningContainers(ch chan<- Containers, wg *sync.WaitGroup, errCh chan error) {
	defer wg.Done()

	var containerSlice Containers
	containers, err := d.client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		errCh <- err
	}

	for _, container := range containers {
		details, err := d.client.InspectContainer(container.ID)
		if err != nil {
			errCh <- err
		}
		containerSlice = append(containerSlice, details)
	}

	ch <- containerSlice
}
