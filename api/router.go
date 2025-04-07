package api

import (
	api_rules "github.com/wavix/w-alerts/api/rules"
	api_status "github.com/wavix/w-alerts/api/status"
	"github.com/wavix/w-alerts/rule"

	"github.com/gin-gonic/gin"
)

type Controllers struct {
	statusController api_status.StatusController
	rulesController  api_rules.RulesController
}

func NewControllers(register *rule.Registry) *Controllers {
	return &Controllers{
		statusController: api_status.NewController(register),
		rulesController:  api_rules.NewController(register),
	}
}

func (controllers *Controllers) Routes(router *gin.Engine) {
	routes := router.Group("/")
	routes.GET("/status", controllers.statusController.GetStatus)
	routes.POST("/api/rules", controllers.rulesController.AddRule)
	routes.PATCH("/api/rules", controllers.rulesController.UpdateRule)
}
