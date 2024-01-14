package main

import (
	"fmt"
	"log"
)

func main() {
	store, err := NewPostgresStore()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v\n", store)

	// server := NewApiServer(":3000", store)
	// server.Run()
}
