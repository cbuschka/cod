package inventory

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"time"
)

type ContainerConfig struct {
	Version       string        `yaml:"version"`
	Path          string        `yaml:"path"`
	ImageName     string        `yaml:"image"`
	ContainerPort int           `yaml:"port"`
	MaxIdleTime   time.Duration `yaml:"maxIdleTime" default:"30s"`
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

	if "cod:config/v1" != containerConfig.Version {
		return nil, fmt.Errorf("unsupported version")
	}

	return &containerConfig, nil
}
