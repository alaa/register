package main

import (
	"log"
	"sync"
	"time"

	"./consul"
	"./docker"
)

const interval time.Duration = 2 * time.Second

func main() {
	var wg sync.WaitGroup

	errCh := make(chan error, 1)
	ready := make(chan struct{}, 1)
	dockerCh := make(chan docker.Containers, 1)
	consulCh := make(chan consul.Services, 1)

	d := docker.New()
	c := consul.New()
	tick := time.Tick(interval)

	for {
		select {
		case <-tick:
			wg.Add(2)
			go d.ListRunningContainers(dockerCh, &wg, errCh)
			go c.Services(consulCh, &wg, errCh)
			wg.Wait()
			ready <- struct{}{}

		case _ = <-ready:
			go resync(dockerCh, consulCh)

		case e := <-errCh:
			log.Printf("Error channel: %s", e)
		}
	}
}

func resync(dockerCh <-chan docker.Containers, consulCh <-chan consul.Services) {
	containers := <-dockerCh
	services := <-consulCh

	c := consul.New()

	go func() {
		for _, container := range containers {
			if !consul.Lookup(container.ID, services) {
				log.Println("registering ", container.Name)
				if err := c.Register(container.ID, container.Name); err != nil {
					log.Println(err)
				}
			}
		}
	}()

	go func() {
		for _, service := range services {
			if !docker.Lookup(service.ID, containers) {
				log.Println("deregistering service ", service.ID)
				c.Deregister(service.ID)
			}
		}
	}()
}
