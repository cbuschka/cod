package engine

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

type ContainerRuntime interface {
	PullImage(ctx context.Context, imageName string) error
	ListContainers(ctx context.Context) ([]types.Container, error)
	StartContainer(ctx context.Context, containerId string) error
	StopContainer(ctx context.Context, containerId string) error
	KillContainer(ctx context.Context, containerId string) error
	InspectContainer(ctx context.Context, containerId string) (*types.ContainerJSON, error)
	CreateContainer(ctx context.Context, containerConfig *container.Config, hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig, platform *specs.Platform, containerName string) (*container.ContainerCreateCreatedBody, error)
}
