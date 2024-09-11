package config

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	LogConf           logx.LogConf
	IncludeNamespaces []string `json:",optional"`
	ExcludeNamespaces []string `json:",optional"`
}

var (
	Path              = "/mutate"
	DefaultNamespace  = "default"
	IncludeNamespaces = []string{
		"*",
	}
	ExcludeNamespaces = []string{
		"kube-node-lease",
		"kube-public",
		"kube-system",
	}
	HostMap = map[string]string{
		"docker.io":       "docker.ketches.cn",
		"registry.k8s.io": "k8s.ketches.cn",
		"k8s.gcr.io":      "k8s-gcr.ketches.cn",
		"ghcr.io":         "ghcr.ketches.cn",
		"quay.io":         "quay.ketches.cn",
		"gcr.io":          "gcr.ketches.cn",
	}
)
