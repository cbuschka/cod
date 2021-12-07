package engine

import (
	"bufio"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	log "github.com/sirupsen/logrus"
	"time"
)

func NewContainerRuntime() (ContainerRuntime, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return ContainerRuntime(&Docker{client: dockerClient}), nil
}

type Docker struct {
	client *client.Client
}

func (docker *Docker) StartContainer(ctx context.Context, containerId string) error {

	err := docker.client.ContainerStart(ctx, containerId, types.ContainerStartOptions{})
	return err
}

func (docker *Docker) CreateContainer(ctx context.Context, containerConfig *container.Config, hostConfig *container.HostConfig,
	networkingConfig *network.NetworkingConfig, platform *specs.Platform, containerName string) (*container.ContainerCreateCreatedBody, error) {

	resp, err := docker.client.ContainerCreate(ctx, containerConfig, hostConfig, networkingConfig, platform, containerName)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (docker *Docker) InspectContainer(ctx context.Context, containerId string) (*types.ContainerJSON, error) {
	info, err := docker.client.ContainerInspect(ctx, containerId)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (docker *Docker) StopContainer(ctx context.Context, containerId string) error {
	timeout := 10 * time.Second
	err := docker.client.ContainerStop(ctx, containerId, &timeout)
	return err
}

func (docker *Docker) KillContainer(ctx context.Context, containerId string) error {
	err := docker.client.ContainerKill(ctx, containerId, "KILL")
	return err
}

func (docker *Docker) ListContainers(ctx context.Context) ([]types.Container, error) {
	containerList, err := docker.client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	return containerList, nil
}

func (docker *Docker) PullImage(ctx context.Context, imageName string) error {
	reader, err := docker.client.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		log.Info(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return nil
}
