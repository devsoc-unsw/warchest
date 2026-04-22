package main

import (
	"log"
	"os"

	. "github.com/tigerbeetle/tigerbeetle-go"
)

func main() {
	tbAddress := os.Getenv("TIGERBEETLE_ADDRESS")
	
	if len(tbAddress) == 0 {
		tbAddress = "3000"
	}

	client, err := NewClient(ToUint128(0), []string{tbAddress})
	
	if err != nil {
		log.Printf("Error creating client: %s", err)
		return
	}
	
	defer client.Close()
}
