package main

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/assert"
	"github.com/wavix/w-alerts/rule"
	"github.com/wavix/w-alerts/utils"
)

func TestElasticSearchQueryWithTerm(t *testing.T) {
	inputStr := `
	{
		"query": {
			"bool": {
				"must": [
					{ "term": { "url.keyword": "/v1/endpoint" } }
				]
			}
		}
	}`

	outputJSON := `
	{
	  "query": {
	    "bool": {
	      "must": [
	        {
	          "term": {
	            "url.keyword": "/v1/endpoint"
	          }
	        },
	        {
	          "range": {
	            "@timestamp": {
	              "gte": "now-1m"
	            }
	          }
	        }
	      ]
	    }
	  }
	}`

	parsed, err := utils.JSONToMap(inputStr)
	if err != nil {
		t.Error(err)
	}

	rule := rule.Rule{
		Period: "1m",
		Request: rule.RuleRequest{
			Elastic: parsed,
		},
	}

	err = rule.AddElasticTimestampCondition()
	if err != nil {
		t.Error(err)
	}

	output, err := json.Marshal(rule.Request.Elastic)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, string(output), utils.JsonFormat(outputJSON))
}

func TestElasticSearchQueryMatchPhrase(t *testing.T) {
	inputStr := `
	{
		"query": {
			"match_phrase": {
            	"url": "match/phrase"
          	}
		}
	}`

	outputJSON := `
	{
	  "query": {
	    "bool": {
	      "must": [
		  	{
	          "range": {
	            "@timestamp": {
	              "gte": "now-2m"
	            }
	          }
	        },
	        {
	          "match_phrase": {
	            "url": "match/phrase"
	          }
	        }
	      ]
	    }
	  }
	}`

	parsed, err := utils.JSONToMap(inputStr)
	if err != nil {
		t.Error(err)
	}

	rule := rule.Rule{
		Period: "2m",
		Request: rule.RuleRequest{
			Elastic: parsed,
		},
	}

	err = rule.AddElasticTimestampCondition()
	if err != nil {
		t.Error(err)
	}

	output, err := json.Marshal(rule.Request.Elastic)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, string(output), utils.JsonFormat(outputJSON))
}
