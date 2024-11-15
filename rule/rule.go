package rule

import (
	"fmt"
	"github.com/wavix/w-alerts/types"
	"github.com/wavix/w-alerts/utils"
	"path/filepath"
	"sync"
	"time"

	"github.com/wavix/go-lib/logger"
)

type Rule struct {
	UUID         string     `json:"uuid"`
	File         string     `json:"file"`
	LastExecuted *time.Time `json:"last_executed"`
	IsFire       bool       `json:"is_fire"`

	Name        string          `json:"name"`
	Description string          `json:"description"`
	Index       string          `json:"index"`
	Period      string          `json:"period"`
	Interval    string          `json:"interval"`
	Request     RuleRequest     `json:"request"`
	Rules       []RuleCondition `json:"rules"`

	RulesResults []interface{} `json:"rules_results"`
}

type RuleRequest struct {
	Elastic map[string]interface{} `json:"elastic"`
	Http    *HttpRequest           `json:"http"`
}

type RuleCondition struct {
	Field    string      `json:"field"`
	Field2   string      `json:"field2"` // If exists, then it's a ratio: Field/Field2 for elastic request
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Status   int         `json:"status"`
}

type Registry struct {
	Rules map[string]*Rule
	Mutex sync.RWMutex
}

type HttpRequest struct {
	Url     string
	Method  *string
	Headers *map[string]string
	Body    *map[string]string
}

type ToggleFire struct {
	IsFire       bool
	Response     types.RuleResponse
	Extra        logger.ExtraData
	RulesResults []interface{}
}

func (rule *Rule) AddElasticTimestampCondition() error {
	rangeTime := fmt.Sprintf("now-%s", rule.Period)

	rangeCondition := map[string]interface{}{
		"range": map[string]interface{}{
			"@timestamp": map[string]interface{}{
				"gte": rangeTime,
			},
		},
	}

	query, ok := rule.Request.Elastic["query"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("query is not map[string]interface{}")
	}
	boolQuery, ok := query["bool"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("bool is not a map[string]interface{}")
	}
	must, ok := boolQuery["must"].([]interface{})
	if !ok {
		return fmt.Errorf("must is not []interface{}")
	}

	must = append(must, rangeCondition)
	boolQuery["must"] = must

	return nil
}

func (rule *Rule) GetIndex() string {
	index := rule.Index

	if index[len(index)-1] == '*' {
		index = index[:len(index)-1] + time.Now().Format("2006.01.02")
	}

	return index
}

func (rule *Rule) ProcessResponse(response types.RuleResponse) {
	conditionsCount := len(rule.Rules)
	triggeredConditions := 0

	if len(rule.Rules) > 0 {

		rulesResults := make([]interface{}, conditionsCount)
		extraData := logger.ExtraData{}

		for index, ruleCondition := range rule.Rules {
			var value interface{}

			if ruleCondition.Field != "" {
				var ok bool
				value, ok = utils.GetValueFromMap(response, ruleCondition.Field)

				if !ok {
					utils.Logger.Context(rule.Name).Error().Msgf("Error getting value for field '%s' from response: %v", response, ruleCondition.Field)
					continue
				}

				if !ok {
					utils.Logger.Context(rule.Name).Error().Msgf("Error converting value to number: %v", value)
					continue
				}

				// If there is a second field, then it's a ratio
				// Field/Field2
				if ruleCondition.Field2 != "" {
					value2, ok := utils.GetValueFromMap(response, ruleCondition.Field2)
					if !ok {
						utils.Logger.Context(rule.Name).Error().Msgf("Error getting value for field2 '%s' from response: %v", response, ruleCondition.Field2)
						continue
					}

					valueNumber2 := utils.ToNumber(value2)
					if !ok {
						utils.Logger.Context(rule.Name).Error().Msgf("Error converting value to number: %v", value)
						continue
					}

					// round to 2 decimal places
					val := utils.ToNumber(value)
					value = float64(int((val/valueNumber2)*100)) / 100
					if valueNumber2 == 0 {
						value = 0
					}
				}
			}

			// "status" field check
			if ruleCondition.Status != 0 {
				value = utils.ToNumber(response["status"])

				if value != utils.ToNumber(ruleCondition.Status) {
					rulesResults[index] = value
					rule.ToggleFire(ToggleFire{IsFire: true, Response: response, Extra: extraData, RulesResults: rulesResults})
					return
				}
			}

			// eq's fields check
			if ruleCondition.Operator == "lt" && utils.ToNumber(value) < utils.ToNumber(ruleCondition.Value) {
				triggeredConditions += 1
			}

			if ruleCondition.Operator == "gt" && utils.ToNumber(value) > utils.ToNumber(ruleCondition.Value) {
				triggeredConditions += 1
			}

			if ruleCondition.Operator == "eq" && value != ruleCondition.Value {
				triggeredConditions += 1
			}

			ruleId := fmt.Sprintf("condition_%d", index+1)
			extraData[ruleId] = value
			rulesResults[index] = value
		}

		isFire := triggeredConditions == conditionsCount
		rule.ToggleFire(ToggleFire{
			IsFire:       isFire,
			Response:     response,
			Extra:        extraData,
			RulesResults: rulesResults,
		})
	}
}

func (rule *Rule) ToggleFire(params ToggleFire) {
	now := time.Now()
	isStatusChanged := false

	rule.LastExecuted = &now

	if params.IsFire != rule.IsFire {
		isStatusChanged = true
	}

	// In the case when the problem is resolved, we need to show the previous statistics in the description.
	// Therefore, we change the list of results only when an isFired event has occurred or the status has not changed
	if !isStatusChanged || (isStatusChanged && params.IsFire) {
		rule.RulesResults = params.RulesResults
	}

	rule.IsFire = params.IsFire

	log := utils.Logger.Context(rule.Name, params.Extra)
	log.Extra("fire", params.IsFire)
	log.Extra("state_changed", isStatusChanged)

	if params.IsFire {
		log.Warn().Msgf("%v", params.Response)
		return
	}

	log.Info().Msgf("%v", params.Response)
}

func (rule *Rule) GetNextRunAt() time.Time {
	if rule.LastExecuted == nil {
		return time.Now().Add(-1 * time.Second)
	}

	duration, err := time.ParseDuration(rule.Interval)
	if err != nil {
		utils.Logger.Context(rule.Name).Error().Msgf("Error parsing duration: %v", err)
		return rule.LastExecuted.Add(1 * time.Minute)
	}

	return rule.LastExecuted.Add(duration).Add(-1 * time.Second)
}

func (rule *Rule) GetRule(path string) error {
	fileName := filepath.Base(path)
	rule.UUID = utils.GenerateRuleUUID(fileName, rule.Name)
	rule.File = path

	if rule.Request.Elastic != nil {
		rule.Request.Elastic["size"] = 0

		err := rule.AddElasticTimestampCondition()
		if err != nil {
			return err
		}
	}

	return nil
}

func (registry *Registry) AddRule(rule Rule) {
	registry.Mutex.Lock()
	defer registry.Mutex.Unlock()

	if registry.Rules[rule.UUID] != nil {
		current := registry.Rules[rule.UUID]

		registry.Rules[rule.UUID] = &rule

		// Preserve the current state
		registry.Rules[rule.UUID].IsFire = current.IsFire
		registry.Rules[rule.UUID].RulesResults = current.RulesResults
		registry.Rules[rule.UUID].LastExecuted = current.LastExecuted
		return
	}

	registry.Rules[rule.UUID] = &rule
}

func (registry *Registry) RemoveRule(uuid string) {
	registry.Mutex.Lock()
	defer registry.Mutex.Unlock()

	delete(registry.Rules, uuid)
}
