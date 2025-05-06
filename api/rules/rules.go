package api_rules

import (
	"net/http"
	"time"

	"github.com/wavix/w-alerts/rule"
	"github.com/wavix/w-alerts/utils"

	"github.com/gin-gonic/gin"
)

type RuleCreationPayload struct {
	UUID        string  `json:"uuid" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Scope       *string `json:"scope"`
	IsFire      bool    `json:"is_fire"`
}

type RuleUpdatePayload struct {
	UUID   string `json:"uuid" binding:"required"`
	IsFire bool   `json:"is_fire"`
}

type RulesController struct {
	registry *rule.Registry
}

func NewController(registry *rule.Registry) RulesController {
	return RulesController{
		registry: registry,
	}
}

func (controller RulesController) AddRule(context *gin.Context) {
	var payload RuleCreationPayload

	if !utils.ValidateBody(context, &payload) {
		return
	}

	now := time.Now().UTC()
	isFire := payload.IsFire

	if _, exists := controller.registry.Rules[payload.UUID]; exists {
		controller.registry.RemoveRule(payload.UUID)
	}

	rule := rule.Rule{
		UUID:          payload.UUID,
		Name:          payload.Name,
		Description:   payload.Description,
		Scope:         payload.Scope,
		LastExecuted:  &now,
		IsFire:        isFire,
		IsStaticAlert: true,
	}

	controller.registry.AddRule(rule)

	logger := utils.Logger.Info()

	if isFire {
		logger = utils.Logger.Warn()
	}

	logger.Msgf("Add static alert (UUID: %s, name: '%s')", payload.UUID, payload.Name)
	context.JSON(http.StatusOK, gin.H{"success": "true", "message": "Rule added successfully", "rule": payload})
}

func (controller RulesController) UpdateRule(context *gin.Context) {
	var payload RuleUpdatePayload

	if !utils.ValidateBody(context, &payload) {
		return
	}

	if _, exists := controller.registry.Rules[payload.UUID]; !exists {
		context.JSON(http.StatusBadRequest, gin.H{"success": "false", "message": "Rule not found"})
		return
	}

	rule := controller.registry.Rules[payload.UUID]
	rule.IsFire = payload.IsFire

	context.JSON(http.StatusOK, gin.H{"success": "true", "message": "Rule successfully updated"})
}
