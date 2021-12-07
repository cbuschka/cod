package main

import "github.com/cbuschka/cod/internal/daemon"

func main() {
	err := daemon.Run()
	if err != nil {
		panic(err)
	}
}
