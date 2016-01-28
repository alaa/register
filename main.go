package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"./docker"
)

const interval time.Duration = 2 * time.Second

func main() {
	var wg sync.WaitGroup

	errCh := make(chan error, 1)
	ready := make(chan struct{}, 1)
	dockerCh := make(chan docker.Containers, 1)

	d := docker.New()
	tick := time.Tick(interval)

	for {
		select {
		case <-tick:
			wg.Add(1)
			go d.ListRunningContainers(dockerCh, &wg, errCh)
			wg.Wait()
			ready <- struct{}{}

		case _ = <-ready:
			resync(dockerCh)

		case x := <-errCh:
			log.Printf("Error: %s", x)
		}
	}
}

func resync(dockerCh <-chan docker.Containers) {
	for _, container := range <-dockerCh {
		fmt.Println(container.ID)
	}
}
