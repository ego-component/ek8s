package ek8s

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"k8s.io/client-go/rest"
)

// Config ...
type Config struct {
	Addr                    string
	Debug                   bool
	Token                   string
	Namespaces              []string
	DeploymentPrefix        string // 命名前缀
	TLSClientConfigInsecure bool
	tokenFile               string
}

// DefaultConfig 返回默认配置，默认采用集群内模式
func DefaultConfig() *Config {
	return &Config{
		Addr:                    inClusterAddr(),
		Token:                   inClusterToken(),
		Namespaces:              []string{inClusterNamespace()},
		TLSClientConfigInsecure: true,
		tokenFile:               tokenFile,
	}
}

func (c *Config) toRestConfig() *rest.Config {
	// 当BearerToken和BearerTokenFile同时不为空时，k8s-go内部为优先采用BearerTokenFile在底层周期刷新token
	// 详见 https://github.com/kubernetes/client-go/blob/v0.24.0/transport/round_trippers.go#L52
	cfg := &rest.Config{
		Host:            c.Addr,
		BearerToken:     c.Token,
		BearerTokenFile: c.tokenFile,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: c.TLSClientConfigInsecure,
		},
	}
	return cfg
}

func inClusterAddr() string {
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT")
	if host == "" || port == "" {
		return ""
	}
	return fmt.Sprintf("https://%s:%s", host, port)
}

const (
	tokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	nsFile    = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

func inClusterToken() string {
	t, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(t))
}

func inClusterNamespace() string {
	t, err := ioutil.ReadFile(nsFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(t))
}
