package main

import (
	"fmt"
	"log"
	"os"

	. "github.com/tigerbeetle/tigerbeetle-go"
)

func init_tb() (Client, error) {
	tbAddress := os.Getenv("TB_ADDRESS")
	if len(tbAddress) == 0 {
		tbAddress = "3000"
	}
	client, err := NewClient(ToUint128(0), []string{tbAddress})
	if err != nil {
		log.Printf("Error creating client: %s", err)
		return nil, err
	}
	return client, nil
}

func main() {
	client, err := init_tb()
	if err != nil {
		log.Printf("Error creating client: %s", err)
		return nil, err
	}
	defer client.Close()
	fmt.Println("Hello, World!")
}
