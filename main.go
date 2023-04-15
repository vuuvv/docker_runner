package main

import (
	"github.com/vuuvv/docker-runner/controllers"
	"github.com/vuuvv/docker-runner/services"
	"github.com/vuuvv/docker-runner/utils"
	"vuuvv.cn/unisoftcn/orca"
	"vuuvv.cn/unisoftcn/orca/auth"
	"vuuvv.cn/unisoftcn/orca/server"
)

func main() {
	app := orca.NewApplication()

	services.DockerService.WatchTask()

	authMiddleware := server.MiddlewareJwt(
		utils.GetAuthConfig(),
		auth.NoAuthorization{},
	)
	app.Default(authMiddleware).Mount(
		controllers.DockerController,
	).Start()
}
