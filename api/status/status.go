package status

import (
	"alerts/rule"
	"alerts/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatusController struct {
	registry *rule.Registry
}

type RuleStatus struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func NewController(registry *rule.Registry) StatusController {
	return StatusController{
		registry: registry,
	}
}

func (controller StatusController) GetStatus(context *gin.Context) {
	response := make([]RuleStatus, 0)

	for _, rule := range controller.registry.Rules {
		status := "ok"

		if rule.IsFire {
			status = "problem"
		}

		description := utils.ReplacePlaceholders(rule.Description, rule.RulesResults)

		response = append(response, RuleStatus{
			UUID:        rule.UUID,
			Name:        rule.Name,
			Description: description,
			Status:      status,
		})
	}

	context.JSON(http.StatusOK, gin.H{"status": response})
}
