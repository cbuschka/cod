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

			routes := engine.getRoutes()
			for _, route := range routes {
				if route.containerInstance != nil && route.lastHit.Before(time.Now().Add(time.Duration(-30)*time.Second)) {
					_ = engine.shutdownRoute(route)
				}
			}
		}
	}()
}
