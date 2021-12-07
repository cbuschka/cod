package engine

import (
	"context"
	"fmt"
	"github.com/cbuschka/cod/internal/inventory"
	"github.com/cbuschka/go-ant-pattern"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/phayes/freeport"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

type Route struct {
	pathPattern       *ant_pattern.AntPattern
	containerInstance *ContainerInstance
	config            *inventory.ContainerConfig
	lastHit           time.Time
}

type ContainerInstance struct {
	id       string
	endpoint ContainerEndpoint
}

type ContainerEndpoint struct {
	Address string
	Port    int
}

type Engine struct {
	sessionId      string
	containerIdSeq *Counter
	routes         []*Route
	configs        map[string]inventory.ContainerConfig
	dockerClient   *client.Client
	janitor        *Janitor
}

func NewEngine() (*Engine, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	sessionId := randSeq(8)
	containerIdSeq := NewCounter()

	janitor := &Janitor{}
	engine := &Engine{routes: []*Route{}, configs: map[string]inventory.ContainerConfig{}, dockerClient: dockerClient, containerIdSeq: containerIdSeq, sessionId: sessionId, janitor: janitor}
	janitor.Start(engine)

	return engine, nil
}

func (engine *Engine) CleanUp(ctx context.Context) error {

	log.Info("Cleaning up...")

	containerList, err := engine.dockerClient.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	for _, container_ := range containerList {
		_, found := container_.Labels["cod:managed"]
		if found {
			log.Infof("Removing container %s...", container_.Names[0])

			err := engine.dockerClient.ContainerKill(ctx, container_.ID, "KILL")
			if err != nil {
				return err
			}
		}
	}

	log.Info("Clean up done.")

	return nil
}

func (engine *Engine) findRoute(path string) (*Route, string, error) {

	var best *Route = nil

	for _, route := range engine.routes {
		if route.pathPattern.Matches(path) && (best == nil || best.pathPattern.Specificity() < route.pathPattern.Specificity()) {
			best = route
		}
	}

	if best != nil {
		groups := best.pathPattern.FindStringSubmatch(path)
		if len(groups) > 1 {
			return best, groups[1], nil
		}
		return best, path, nil
	}

	return nil, path, fmt.Errorf("route not found")
}

func (engine *Engine) GetOrStartContainer(path string) (*ContainerEndpoint, string, error) {

	route, downstreamPath, err := engine.findRoute(path)
	if err != nil {
		return nil, "", err
	}

	log.Infof("%s matches route pattern %s, target path %s", path, route.pathPattern.String(), downstreamPath)

	if route.containerInstance != nil {
		route.lastHit = time.Now()
		return &route.containerInstance.endpoint, downstreamPath, nil
	}

	containerInstance, err := engine.StartContainer(route.config)
	if err != nil {
		return nil, "", err
	}

	route.containerInstance = containerInstance
	route.lastHit = time.Now()

	return &containerInstance.endpoint, downstreamPath, nil
}

func (engine *Engine) StartContainer(config *inventory.ContainerConfig) (*ContainerInstance, error) {

	containerName := fmt.Sprintf("cod_%s_%d", engine.sessionId, engine.containerIdSeq.Next())

	log.Infof("Starting container %s (image=%s)...", containerName, config.ImageName)

	ctx := context.Background()

	reader, err := engine.dockerClient.ImagePull(ctx, config.ImageName, types.ImagePullOptions{})
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader)

	networkingConfig := network.NetworkingConfig{EndpointsConfig: make(map[string]*network.EndpointSettings)}
	hostConfig := container.HostConfig{
		AutoRemove: true,
		Resources: container.Resources{
			CPUShares: 10,
			Memory:    config.MemoryBytes},
	}
	hostConfig.PortBindings = make(nat.PortMap)
	containerPort, err := nat.NewPort("tcp", fmt.Sprintf("%d", config.ContainerPort))
	if err != nil {
		return nil, err
	}

	hostPort := config.HostPort
	if hostPort == 0 {
		hostPort, err = freeport.GetFreePort()
		if err != nil {
			return nil, err
		}
	}

	hostConfig.PortBindings[containerPort] = []nat.PortBinding{{HostPort: fmt.Sprintf("%d/tcp", hostPort), HostIP: config.HostAddress}}
	labels := make(map[string]string)
	labels["cod:managed"] = "true"
	labels["cod:configFilename"] = config.Filename
	containerConfig := container.Config{
		Tty:    false,
		Image:  config.ImageName,
		Labels: labels,
	}
	resp, err := engine.dockerClient.ContainerCreate(ctx, &containerConfig, &hostConfig, &networkingConfig, nil, containerName)
	if err != nil {
		return nil, err
	}

	err = engine.dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}

	// mappedPort, err := engine.getExposedPort(ctx, config.ContainerPort, resp.ID)
	err = waitForAvailableViaHttp(config.HostAddress, hostPort)
	if err != nil {
		return nil, err
	}

	return &ContainerInstance{id: resp.ID, endpoint: ContainerEndpoint{Address: config.HostAddress, Port: hostPort}}, nil
}

func (engine *Engine) getExposedPort(ctx context.Context, containerPort int, containerId string) (int, error) {

	info, err := engine.dockerClient.ContainerInspect(ctx, containerId)
	if err != nil {
		return -1, err
	}

	for port, portBindings := range info.HostConfig.PortBindings {
		if port.Int() == containerPort {
			hostPort, err := nat.ParsePort(portBindings[0].HostPort)
			if err != nil {
				return -1, err
			}
			return hostPort, nil
		}
	}

	return -1, fmt.Errorf("no binding found")
}

func (engine *Engine) getRoutes() []*Route {
	return engine.routes
}

func (engine *Engine) AddContainerConfig(containerConfig *inventory.ContainerConfig) error {

	pathPattern, err := ant_pattern.ParseAntPattern(containerConfig.Path)
	if err != nil {
		return err
	}

	route := Route{pathPattern: pathPattern, containerInstance: nil, config: containerConfig, lastHit: time.Now()}
	engine.routes = append(engine.routes, &route)

	return nil
}

func (engine *Engine) shutdownRoute(route *Route) error {

	log.Infof("Shutting down %s...", route.pathPattern)

	containerInstance := route.containerInstance
	route.containerInstance = nil

	if containerInstance != nil {
		ctx := context.TODO()
		err := engine.dockerClient.ContainerKill(ctx, containerInstance.id, "KILL")
		if err != nil {
			return err
		}
	}

	return nil
}
