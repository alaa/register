package resync

import (
	"log"
	"sync"
	"time"

	"../consul"
	"../docker"
)

type Register struct {
	dockerAgent *docker.Docker
	dockerCh    chan docker.Containers
	consulAgent *consul.Consul
	consulCh    chan consul.Services
}

func New() *Register {
	return &Register{
		dockerAgent: docker.New(),
		dockerCh:    make(chan docker.Containers),
		consulAgent: consul.New(),
		consulCh:    make(chan consul.Services),
	}
}

func (r *Register) Read(wg sync.WaitGroup) {
	for {
		select {
		case <-time.After(5 * time.Second):
			wg.Add(2)
			go r.dockerAgent.ListRunningContainers(r.dockerCh, &wg)
			go r.consulAgent.Services(r.consulCh, &wg)
			wg.Wait()
		}
	}
}

func (r *Register) Update() {
	for {
		containers, services := <-r.dockerCh, <-r.consulCh
		r.register(containers, services)
		r.deregister(containers, services)
	}
}

func (r *Register) register(containers docker.Containers, services consul.Services) {
	for _, container := range containers {
		if !consul.Lookup(container.ID, services) {
			log.Println("registering ", container.Name)
			if err := r.consulAgent.Register(container.ID, container.Name); err != nil {
				log.Println(err)
			}
		}
	}
}

func (r *Register) deregister(containers docker.Containers, services consul.Services) {
	for _, service := range services {
		if !docker.Lookup(service.ID, containers) {
			log.Println("deregistering service ", service.ID)
			r.consulAgent.Deregister(service.ID)
		}
	}
}
