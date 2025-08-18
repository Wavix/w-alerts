package api_status

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/wavix/w-alerts/rule"
	"github.com/wavix/w-alerts/utils"

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
	Severity    string `json:"severity"`
}

func NewController(registry *rule.Registry) StatusController {
	return StatusController{
		registry: registry,
	}
}

func (controller StatusController) GetStatus(context *gin.Context) {
	response := make([]RuleStatus, 0)

	for _, r := range controller.registry.Rules {
		status := "ok"

		if r.IsFire {
			status = "problem"
		}

		description := utils.ReplacePlaceholders(r.Description, r.RulesResults)

		name := r.Name

		if r.Scope != nil && *r.Scope != "" {
			name = fmt.Sprintf("[%s] %s", strings.ToUpper(*r.Scope), r.Name)
		}

		severity := rule.RuleSeverityInfo
		if r.Severity != nil {
			severity = *r.Severity
		}

		response = append(response, RuleStatus{
			UUID:        r.UUID,
			Name:        name,
			Description: description,
			Status:      status,
			Severity:    string(severity),
		})
	}

	context.JSON(http.StatusOK, gin.H{"status": response})
}
