package main

import (
	"fmt"
	"github.com/bilisound/server/internal/utils"
	"log"
)

func main() {
	fmt.Printf("Hello world!!!\n")

	err := utils.ParseVideo()
	if err != nil {
		log.Fatal(err)
	}
}
