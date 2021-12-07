package inventory

import (
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

type Inventory struct {
	dirname string
}

func NewInventory(dirname string) (*Inventory, error) {
	return &Inventory{dirname: dirname}, nil
}

func (inventory *Inventory) GetConfigs() ([]*ContainerConfig, error) {

	containerConfigs := []*ContainerConfig{}
	err := filepath.Walk(inventory.dirname, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".yml") && !info.IsDir() {
			log.Infof("Loading config from %s...", path)

			containerConfig, err := LoadContainerConfig(path)
			if err != nil {
				return err
			}

			containerConfigs = append(containerConfigs, containerConfig)
		}

		return nil
	})

	return containerConfigs, err
}
