package main

import (
	"github.com/gin-gonic/gin"
	"github.com/vuuvv/docker-runner/controllers"
	"github.com/vuuvv/docker-runner/services"
	"github.com/vuuvv/orca"
	"github.com/vuuvv/orca/server"
)

func main() {
	app := orca.NewApplication()

	services.DockerService.WatchTask()

	authMiddleware := server.MiddlewareJwt(
		app.GetHttpServer().GetConfig(),
		server.NoAuthorization{},
	)
	app.Use(server.MiddlewareId, gin.Logger(), gin.Recovery(), authMiddleware).Mount(
		controllers.DockerController,
	).Start()
}
