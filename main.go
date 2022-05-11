package main

import (
	"github.com/gin-gonic/gin"
	"github.com/vuuvv/docker-runner/controllers"
	"github.com/vuuvv/docker-runner/services"
	"vuuvv.cn/unisoftcn/orca"
	"vuuvv.cn/unisoftcn/orca/auth"
	"vuuvv.cn/unisoftcn/orca/server"
)

func main() {
	app := orca.NewApplication()

	services.DockerService.WatchTask()

	authMiddleware := server.MiddlewareJwt(
		app.GetHttpServer().GetConfig(),
		auth.NoAuthorization{},
	)
	app.Use(server.MiddlewareId, gin.Logger(), gin.Recovery(), authMiddleware).Mount(
		controllers.DockerController,
	).Start()
}
