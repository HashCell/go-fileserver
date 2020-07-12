package main

import (
	"log"

	"github.com/HashCell/go-fileserver/config"
	"github.com/HashCell/go-fileserver/route"
)

func main() {
	router := route.Router()
	err := router.Run(config.UploadServiceHost)
	if err != nil {
		log.Println(err)
	}
}
