package discovery

import (
	"strconv"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
)

type Registrar interface {
	Register(serviceName string, serviceAddr string) error
	DeregisterAll()
}

type registrar struct {
	factory    *consulRegistrarFactory
	registrars []sd.Registrar
}

func NewConsulRegistrar(consulAddr string, logger log.Logger) Registrar {
	return &registrar{
		factory: newConsulRegistrarFactory(consulAddr, logger),
	}
}

func (r registrar) Register(serviceName string, serviceAddr string) error {
	reg, err := r.factory.Create(serviceName, serviceAddr)
	if err != nil {
		return err
	}

	reg.Register()
	r.registrars = append(r.registrars, reg)

	return nil
}

func (r registrar) DeregisterAll() {
	for i := range r.registrars {
		r.registrars[i].Deregister()
	}
}

type consulRegistrarFactory struct {
	consulAddr string
	logger     log.Logger
}

func newConsulRegistrarFactory(consulAddr string, logger log.Logger) *consulRegistrarFactory {
	return &consulRegistrarFactory{
		consulAddr: consulAddr,
		logger:     logger,
	}
}

func (r *consulRegistrarFactory) Create(serviceName string, serviceAddr string) (sd.Registrar, error) {
	var client consulsd.Client
	{
		consulConfig := api.DefaultConfig()
		consulConfig.Address = r.consulAddr
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			_ = r.logger.Log("register error", err)
			return nil, err
		}
		client = consulsd.NewClient(consulClient)
		consulClient.KV()
	}

	check := api.AgentServiceCheck{
		GRPC:     serviceAddr,
		Interval: "10s",
		Timeout:  "1s",
		Notes:    "health check",
	}

	sa := strings.Split(serviceAddr, ":")
	port, _ := strconv.Atoi(sa[1])
	asr := api.AgentServiceRegistration{
		ID:      serviceName,
		Name:    serviceName,
		Address: sa[0],
		Port:    port,
		Check:   &check,
	}

	return consulsd.NewRegistrar(client, &asr, r.logger), nil
}
