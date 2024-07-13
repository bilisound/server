package server

import (
	"fmt"
	"github.com/bilisound/server/internal/config"
	"github.com/bilisound/server/internal/server/routes"
	"github.com/gin-gonic/gin"
	"log"
)

func Start() {
	gin.SetMode(gin.DebugMode)
	engine := gin.Default()

	routes.InitRoute(engine, "/api/internal")

	log.Println("Starting server...")
	err := engine.Run(fmt.Sprintf("%s:%d", config.Global.MustString("server.host"), config.Global.Int("server.port")))
	if err != nil {
		log.Fatal(err)
	}
}
