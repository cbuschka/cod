package daemon

import (
	"context"
	enginePkg "github.com/cbuschka/cod/internal/engine"
	inventoryPkg "github.com/cbuschka/cod/internal/inventory"
	proxyPkg "github.com/cbuschka/cod/internal/proxy"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
)

func Run() error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	configDir, err := filepath.Abs(filepath.Join(path, "config"))
	if err != nil {
		return err
	}
	log.Infof("Config dir is %s...", configDir)
	inventory, err := inventoryPkg.NewInventory(configDir)
	if err != nil {
		return err
	}

	engine, err := enginePkg.NewEngine()
	if err != nil {
		return err
	}
	err = engine.CleanUp(context.TODO())
	if err != nil {
		return err
	}

	containerConfigs, err := inventory.GetConfigs()
	if err != nil {
		return err
	}
	for _, containerConfig := range containerConfigs {
		err = engine.AddContainerConfig(enginePkg.ContainerConfig{Path: containerConfig.Path,
			ImageName:     containerConfig.ImageName,
			ContainerPort: containerConfig.ContainerPort,
			MaxIdleTime:   containerConfig.MaxIdleTime,
		})
		if err != nil {
			return err
		}
	}

	proxy, err := proxyPkg.NewProxy(engine)
	if err != nil {
		return err
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {

		log.Infof("Got a request to %s...", request.URL.Path)

		err := proxy.ForwardToContainer(writer, request)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	return http.ListenAndServe(":8080", nil)
}
