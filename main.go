package main

import (
	"fmt"
	"github.com/ScottMaclure/proglog/internal/server"
)

func main() {
	log := server.NewLog()
	fmt.Println("Log:", log)
}
