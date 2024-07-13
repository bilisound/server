package main

import (
	"github.com/bilisound/server/internal/utils"
	"log"
)

func main() {
	err := utils.ParseVideo()
	if err != nil {
		log.Fatal(err)
	}
}
