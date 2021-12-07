package engine

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type Janitor struct {
}

func (janitor *Janitor) Start(engine *Engine) {
	log.Info("Janitor started...")

	go func() {
		for {
			time.Sleep(1 * time.Second)

			instances := engine.getContainerInstances()
			for _, instance := range instances {
				if instance.lastHit.Before(time.Now().Add(time.Duration(-10) * time.Second)) {
					_ = engine.shutdownContainerInstance(instance)
				}
			}
		}
	}()
}
