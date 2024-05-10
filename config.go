package ek8s

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"k8s.io/client-go/rest"
)

// Config 定义了ek8s组件配置结构
type Config struct {
	// Addr k8s API Server 地址
	Addr string
	// Debug 是否开启debug模式
	Debug bool
	// Token k8s API Server 请求token
	Token string
	// Token k8s API Server 请求token file
	// 本地运行时：一定需要显式设置 tokenFile = ""
	// 集群模式下运行时：需要 tokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token" 或释掉 tokenFile 这个 key
	TokenFile string
	// Namespaces 需要进行查询和监听的 Namespace 列表
	Namespaces []string
	// DeploymentPrefix 命名前缀
	DeploymentPrefix string
	// TLSClientConfigInsecure 是否启用 TLS
	TLSClientConfigInsecure bool
}

// DefaultConfig 返回默认配置，默认采用集群内模式
func DefaultConfig() *Config {
	return &Config{
		Addr:                    inClusterAddr(),
		Token:                   inClusterToken(),
		Namespaces:              []string{inClusterNamespace()},
		TLSClientConfigInsecure: true,
		// NOTICE, 如配置文件中"tokenFile"这个key不存在，则TokenFile值为tokenFile常量
		// 如key存在，无论这个值是否为空，则TokenFile值都为该key的值
		TokenFile: tokenFile,
	}
}

func (c *Config) toRestConfig() *rest.Config {
	// 当BearerToken和BearerTokenFile同时不为空时，k8s-go内部为优先采用BearerTokenFile在底层周期刷新token
	// 详见 https://github.com/kubernetes/client-go/blob/v0.24.0/transport/round_trippers.go#L52
	cfg := &rest.Config{
		Host:            c.Addr,
		BearerToken:     c.Token,
		BearerTokenFile: c.TokenFile,
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
