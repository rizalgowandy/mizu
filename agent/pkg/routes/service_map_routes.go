package routes

import (
	"github.com/up9inc/mizu/agent/pkg/controllers"

	"github.com/gin-gonic/gin"
)

func ServiceMapRoutes(ginApp *gin.Engine) {
	routeGroup := ginApp.Group("/servicemap")

	controller := controllers.NewServiceMapController()

	routeGroup.GET("/status", controller.Status)
	routeGroup.GET("/get", controller.Get)
	routeGroup.GET("/reset", controller.Reset)
}
