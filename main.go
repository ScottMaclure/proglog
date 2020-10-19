package main

import (
	"fmt"
	"internal/server/log"
)

func main() {
	log := log.NewLog()
	fmt.Println("Log:", log)
}
