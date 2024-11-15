package api

import (
	"alerts/api/status"
	"alerts/rule"

	"github.com/gin-gonic/gin"
)

type Controllers struct {
	statusController status.StatusController
}

func NewControllers(register *rule.Registry) *Controllers {
	return &Controllers{
		statusController: status.NewController(register),
	}
}

func (controllers *Controllers) Routes(router *gin.Engine) {
	routes := router.Group("/")
	routes.GET("/status", controllers.statusController.GetStatus)
}
