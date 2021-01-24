package config

import (
	"net/http"

	"github.com/hashicorp/consul/api"
)

type KV interface {
	Get(key string) ([]byte, error)
}

type consulKV struct {
	kv *api.KV
}

func NewConsulKV(consulAddr string) (KV, error) {
	client, err := api.NewClient(&api.Config{
		Address:    consulAddr,
		HttpClient: http.DefaultClient,
	})
	if err != nil {
		return nil, err
	}

	return &consulKV{
		client.KV(),
	}, nil
}

func (r *consulKV) Get(key string) ([]byte, error) {
	pair, _, err := r.kv.Get(key, nil)
	if err != nil {
		return nil, err
	}

	return pair.Value, nil
}
