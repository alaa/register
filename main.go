package main

import (
	"sync"

	"./consul"
	"./docker"
	"./resync"
)

func main() {
	var wg sync.WaitGroup
	dockerCh := make(chan docker.Containers)
	consulCh := make(chan consul.Services)
	dockerAgent := docker.New()
	consulAgent := consul.New()

	go resync.Read(dockerAgent, dockerCh, consulAgent, consulCh, wg)
	resync.Write(dockerCh, consulCh, consulAgent)
}
