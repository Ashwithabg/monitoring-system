package main

import (
	"distributed-example/src/distributed/coordinator"
	"fmt"
)

func main() {
	ql := coordinator.NewQueueListener()
	go ql.ListenForNewSource()

	var a string
	fmt.Scanln(&a)
}
