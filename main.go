package main

import (
	"github.com/gin-gonic/gin"
	"test/awesomeproject/conf"
	"test/awesomeproject/service"
	"test/awesomeproject/web/handler"
)

var (
	svcBenchmarkCenter = "BenchmarkCenter"
)

func main() {
	var cfg conf.Config
	conf.ReadConfig(&cfg)

	err := service.Mgr().Register(
		service.NewBenchmarkCenter(svcBenchmarkCenter, cfg),
	)
	if err != nil {
		return
	}

	r := gin.Default()
	//api := r.Group("/api")

	SetupApiRouters(r)
	err = r.Run(cfg.Server.Host + ":" + cfg.Server.Port)
	if err != nil {
		return
	}
}

func SetupApiRouters(api *gin.Engine) {
	api.GET("/sites", handler.GetSitesData)
}
