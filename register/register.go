package register

import (
	"log"
	"regexp"
	"strings"
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

func (r *Register) Read() {
	var wg sync.WaitGroup
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
	var wg sync.WaitGroup
	for {
		select {
		default:
			containers, services := <-r.dockerCh, <-r.consulCh
			wg.Add(2)
			go r.register(containers, services, &wg)
			go r.deregister(containers, services, &wg)
			wg.Wait()
		}
	}
}

func (r *Register) register(containers docker.Containers, services consul.Services, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, container := range containers {
		if !consul.Lookup(container.ID, services) {

			log.Printf("ID %v", container.ID)
			log.Printf("Name %v", container.Name)
			log.Printf("Ports %v", container.Config.ExposedPorts)
			log.Printf("Env %v", serviceVars(container.Config.Env))

			if err := r.consulAgent.Register(container.ID, container.Name); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func (r *Register) deregister(containers docker.Containers, services consul.Services, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, service := range services {
		if !docker.Lookup(service.ID, containers) {
			log.Println("deregistering service ", service.ID)
			r.consulAgent.Deregister(service.ID)
		}
	}
}

type envvar map[string]string

var envKeyPattern = regexp.MustCompile(`^SERVICE_[a-zA-Z0-9]{3,}$`)
var envValuePattern = regexp.MustCompile(`^[a-zA-Z0-9]{3,}$`)

func valid(str string, pattern *regexp.Regexp) bool {
	if pattern.MatchString(str) {
		return true
	}
	return false
}

func keyPairs(vars []string) envvar {
	env := make(envvar)
	for _, v := range vars {
		if parts := strings.Split(v, "="); len(parts) > 0 {
			key, value := parts[0], parts[1]
			env[key] = value
		}
	}
	return env
}

func serviceVars(vars []string) envvar {
	serviceVars := make(envvar)
	for key, value := range keyPairs(vars) {
		if valid(key, envKeyPattern) && valid(value, envValuePattern) {
			serviceVars[key] = value
		}
	}
	return serviceVars
}
