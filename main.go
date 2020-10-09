package main

import (
	"github.com/tarikhagustia/wp_node_go/kernel"
	"log"
)

func main() {
	er := kernel.Initialize()
	if er != nil {
		log.Fatalln("Application error while starting")
	}
}
