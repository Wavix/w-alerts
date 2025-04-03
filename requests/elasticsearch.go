package requests

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/wavix/w-alerts/rule"
	"github.com/wavix/w-alerts/types"
)

type ElasticResponse struct {
	Aggregations map[string]interface{}    `json:"aggregations,omitempty"`
	Hits         *Hits                     `json:"hits,omitempty"`
}

type Hits struct {
    Total *Total                   `json:"total,omitempty"`
    Hits  []map[string]interface{} `json:"hits,omitempty"`
}

type Total struct {
	Value int `json:"value,omitempty"`
}

var esTransport = &http.Transport{
	TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	DisableKeepAlives: false,
}

var esClient = &http.Client{Transport: esTransport, Timeout: 10 * time.Second}

type ResultWithAggregations struct {
	Aggregations map[string]interface{} `json:"aggregations"`
	Value        *int                   `json:"value"`
}

func ExecElasticRule(rule *rule.Rule) (types.RuleResponse, error) {
	if rule.Request.Elastic == nil {
		return nil, errors.New("rule does not have an elastic")
	}

	jsonData, err := json.Marshal(rule.Request.Elastic)
	if err != nil {
		return nil, err
	}

	index := rule.GetIndex()
	url := fmt.Sprintf("https://%s:%s/%s/_search", os.Getenv("ES_HOST"), os.Getenv("ES_PORT"), index)

	username := os.Getenv("ES_USER")
	password := os.Getenv("ES_PASSWORD")

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(username, password)

	resp, err := esClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		return nil, errors.New("error getting response from ES: " + string(body))
	}

	var response ElasticResponse

	err = json.Unmarshal([]byte(body), &response)
	if err != nil {
		return nil, errors.New("error unmarshalling ES response")
	}

    ResultWithAggregations := ResultWithAggregations{
        Aggregations: response.Aggregations,
        Value:        &response.Hits.Total.Value,
    }

    if len(response.Aggregations) > 0 {
        ResultWithAggregations.Aggregations = response.Aggregations
    }

	result, err := structToMap(ResultWithAggregations)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func structToMap(object interface{}) (types.RuleResponse, error) {
	result := make(map[string]interface{})

	data, err := json.Marshal(object)
	if err != nil {
		return nil, errors.New("error marshalling result")
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, errors.New("error unmarshalling result")
	}

	return result, nil
}
