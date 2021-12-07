package inventory

import (
	"bytes"
	"fmt"
	"github.com/docker/go-units"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"time"
)

type ContainerConfig struct {
	Filename      string        `yaml:"-"`
	Version       string        `yaml:"version"`
	Path          string        `yaml:"path"`
	ImageName     string        `yaml:"image"`
	ContainerPort int           `yaml:"port"`
	MaxIdleTime   time.Duration `yaml:"maxIdleTime"`
	HostPort      int           `yaml:"hostPort"`
	HostAddress   string        `yaml:"hostAddress"`
	CPUShares     int           `yaml:"cpu"`
	MemoryStr     string        `yaml:"memory"`
	MemoryBytes   int64         `yaml:"-"`
}

func LoadContainerConfig(filename string) (*ContainerConfig, error) {

	yamlBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	containerConfig := ContainerConfig{}
	err = yaml.NewDecoder(bytes.NewReader(yamlBytes)).Decode(&containerConfig)
	if err != nil {
		return nil, err
	}

	containerConfig.Filename = filename

	if "cod:config/v1" != containerConfig.Version {
		return nil, fmt.Errorf("unsupported version")
	}

	if containerConfig.HostAddress == "" {
		containerConfig.HostAddress = "127.0.0.1"
	}

	if containerConfig.MemoryStr == "" {
		containerConfig.MemoryStr = "20MB"
	}

	size, err := units.FromHumanSize(containerConfig.MemoryStr)
	if err != nil {
		return nil, err
	}
	containerConfig.MemoryBytes = size

	if containerConfig.CPUShares == 0 {
		containerConfig.CPUShares = 10
	}

	if containerConfig.MaxIdleTime == 0 {
		containerConfig.MaxIdleTime = 30 * time.Second
	}

	return &containerConfig, nil
}
