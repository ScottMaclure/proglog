package main

import (
	"fmt"
	"github.com/ScottMaclure/proglog/internal/server"
)

func appendLog(log *server.Log, message string) {
	r := server.Record{Value: []byte("Hello, World!")}
	offset, err := log.Append(r)
	if err != nil {
		panic(err)
	}
	fmt.Println("offset:", offset)
	fmt.Println("Log:", log)
}

func main() {
	log := server.NewLog()
	fmt.Println("Log:", log)

	appendLog(log, "Hello, World")
	appendLog(log, "Goodbye, World")
	appendLog(log, "From Hell's heart, I stab at thee...")

}
