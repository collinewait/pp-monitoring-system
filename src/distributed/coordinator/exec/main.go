package main

import (
	"fmt"

	"github.com/collinewait/pp-monitoring-system/src/distributed/coordinator"
)

func main() {
	queueListener := coordinator.NewQueueListener()
	go queueListener.ListenForNewSource()

	var a string
	fmt.Scanln(&a)
}
