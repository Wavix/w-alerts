package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"slices"

	"github.com/wavix/w-alerts/rule"
	"github.com/wavix/w-alerts/utils"

	"path/filepath"
	"strings"
)

func loadRules(registry *rule.Registry) {
	dir := "./rules"

	if os.Getenv("RULES_DIR") != "" {
		dir = os.Getenv("RULES_DIR")
	}

	utils.Logger.Info().Msgf("Loading rules from %v", dir)

	files, err := os.ReadDir(dir)
	if err != nil {
		utils.Logger.Error().Msgf("Error reading rules directory: %v", err)
		os.Exit(1)
	}

	// array of new UUIDs
	newUUIDs := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			rulePath := filepath.Join(dir, file.Name())

			rule, err := loadRule(rulePath)
			if err != nil {
				utils.Logger.Context(file.Name()).Error().Msgf("Error reading rule: %v", err)
				continue
			}

			for _, r := range *rule {
				newUUIDs = append(newUUIDs, r.UUID)
				registry.AddRule(r)

				utils.Logger.Context(r.Name).Info().Msgf("Rule loaded from %v", rulePath)
			}
		}
	}

	// remove rules that are not in the new list
	registry.Mutex.Lock()
	for uuid := range registry.Rules {
		if !slices.Contains(newUUIDs, uuid) {
			delete(registry.Rules, uuid)
		}
	}
	registry.Mutex.Unlock()
}

func loadRule(path string) (*[]rule.Rule, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer jsonFile.Close() // nolint:errcheck

	jsonBytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var rules []rule.Rule
	var rule rule.Rule

	if err = json.Unmarshal(jsonBytes, &rule); err == nil {

		err := rule.GetRule(path)
		if err != nil {
			return nil, err
		}

		rules = append(rules, rule)

	} else if err = json.Unmarshal(jsonBytes, &rules); err == nil {
		for i := range rules {

			err := rules[i].GetRule(path)
			if err != nil {
				return nil, err
			}
		}
	} else {
		return nil, errors.New("error unmarshalling rule")
	}

	return &rules, nil
}
