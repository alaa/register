package resync

import (
	"log"
	"sync"
	"time"

	"../consul"
	"../docker"
)

func Read(dockerAgent *docker.Docker, dockerCh chan docker.Containers, consulAgent *consul.Consul, consulCh chan consul.Services, wg sync.WaitGroup) {
	for {
		select {
		case <-time.After(5 * time.Second):
			wg.Add(2)
			go dockerAgent.ListRunningContainers(dockerCh, &wg)
			go consulAgent.Services(consulCh, &wg)
			wg.Wait()
		}
	}
}

func Write(dockerCh <-chan docker.Containers, consulCh <-chan consul.Services, consulAgent consul.ServiceDiscovery) {
	for {
		containers, services := <-dockerCh, <-consulCh
		register(containers, services, consulAgent)
		deregister(containers, services, consulAgent)
	}
}

func register(containers docker.Containers, services consul.Services, consulAgent consul.ServiceDiscovery) {
	for _, container := range containers {
		if !consul.Lookup(container.ID, services) {
			log.Println("registering ", container.Name)
			if err := consulAgent.Register(container.ID, container.Name); err != nil {
				log.Println(err)
			}
		}
	}
}

func deregister(containers docker.Containers, services consul.Services, consulAgent consul.ServiceDiscovery) {
	for _, service := range services {
		if !docker.Lookup(service.ID, containers) {
			log.Println("deregistering service ", service.ID)
			consulAgent.Deregister(service.ID)
		}
	}
}
