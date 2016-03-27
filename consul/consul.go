package consul

import (
	"fmt"
	"sync"

	consul "github.com/hashicorp/consul/api"
)

type Consul struct {
	agent *consul.Agent
}

type Services map[string]*consul.AgentService

func New() *Consul {
	client, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		panic(err)
	}
	return &Consul{
		agent: client.Agent(),
	}
}

func (c *Consul) Services(consulCh chan Services, wg *sync.WaitGroup) {
	defer wg.Done()
	services, err := c.agent.Services()
	if err != nil {
		fmt.Println(err)
		return
	}

	consulCh <- services
}

func (c *Consul) Register(serviceID, serviceName string) error {
	return c.agent.ServiceRegister(&consul.AgentServiceRegistration{
		ID:   serviceID,
		Name: serviceName,
	})
}

func (c *Consul) Deregister(serviceID string) error {
	return c.agent.ServiceDeregister(serviceID)
}

func Lookup(containerID string, services Services) bool {
	for _, service := range services {
		if containerID == service.ID {
			return true
		}
	}
	return false
}
