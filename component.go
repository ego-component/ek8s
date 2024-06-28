package ek8s

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gotomicro/ego/core/elog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

const PackageName = "component.ek8s"
const defaultRsync = 5 * time.Minute

const (
	KindPods      = "pods"
	KindEndpoints = "endpoints"
)

type (
	ListOptions = metav1.ListOptions
)

// Component ...
type Component struct {
	name   string
	config *Config
	*kubernetes.Clientset
	logger *elog.Component
	locker sync.RWMutex
}

type KubernetesEvent struct {
	IPs       []string
	EventType watch.EventType
}

// New ...
func newComponent(name string, config *Config, logger *elog.Component) *Component {
	// 如果没有开启，那么返回一个nil的k8s client，这个时候，认为业务方不会使用这个组件。
	// 这个目的是为了docker-compose这类业务，虽然使用了同样代码，enable=false，让业务不panic
	if !config.Enable {
		return &Component{
			name:      name,
			config:    config,
			logger:    logger,
			Clientset: nil,
		}
	}
	client, err := kubernetes.NewForConfig(config.toRestConfig())
	if err != nil {
		logger.Panic("new component err", elog.FieldErr(err))
	}
	return &Component{
		name:      name,
		config:    config,
		logger:    logger,
		Clientset: client,
	}
}

func (c *Component) Config() Config {
	return *c.config
}

func (c *Component) ListPods(option ListOptions) (pods []*v1.PodList, err error) {
	pods = make([]*v1.PodList, 0)
	for _, ns := range c.config.Namespaces {
		v1Pods, err := c.CoreV1().Pods(ns).List(context.Background(), option)

		if err != nil {
			return nil, fmt.Errorf("list pods in namespace (%s), err: %w", ns, err)
		}
		pods = append(pods, v1Pods)
	}
	return
}

func (c *Component) ListEndpoints(option ListOptions) (endPoints []*v1.EndpointsList, err error) {
	endPoints = make([]*v1.EndpointsList, 0)
	for _, ns := range c.config.Namespaces {
		v1EndPoints, err := c.CoreV1().Endpoints(ns).List(context.Background(), option)
		if err != nil {
			return nil, fmt.Errorf("list endpoints in namespace (%s), err: %w", ns, err)
		}
		endPoints = append(endPoints, v1EndPoints)
	}
	return
}

func (c *Component) ListPodsByName(name string) (pods []*v1.Pod, err error) {
	pods = make([]*v1.Pod, 0)
	for _, ns := range c.config.Namespaces {
		v1Pods, err := c.CoreV1().Pods(ns).Get(context.Background(), c.getDeploymentName(name), metav1.GetOptions{})

		if err != nil {
			return nil, fmt.Errorf("list pods in namespace (%s), err: %w", ns, err)
		}
		pods = append(pods, v1Pods)
	}
	return
}

func (c *Component) ListEndpointsByName(name string) (endPoints []*v1.Endpoints, err error) {
	endPoints = make([]*v1.Endpoints, 0)
	for _, ns := range c.config.Namespaces {
		v1EndPoints, err := c.CoreV1().Endpoints(ns).Get(context.Background(), c.getDeploymentName(name), metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("list endpoints in namespace (%s), err: %w", ns, err)
		}
		endPoints = append(endPoints, v1EndPoints)
	}
	return
}

func (c *Component) NewWatcherApp(ctx context.Context, appName string, kind string) (app *WatcherApp, err error) {
	app = newWatcherApp(c.Clientset, appName, kind, c.config.DeploymentPrefix, c.logger)
	for _, ns := range c.config.Namespaces {
		err = app.watch(ctx, ns)
		if err != nil {
			return app, err
		}
	}
	return app, nil
}

func (c *Component) getDeploymentName(name string) string {
	return c.config.DeploymentPrefix + name
}
