package topologygo
package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	fmt.Println("Topology service: Dependency graph crawler")
	fmt.Println("This service crawls live dependencies and builds the impact graph.")
	
	port := os.Getenv("TOPOLOGY_PORT")
	if port == "" {
		port = "8081"
	}
	
	log.Printf("Topology service would listen on port %s (not yet implemented)\n", port)
}
