package requests

import (
	"alerts/rule"
	"alerts/types"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

type HttpResults struct {
	Status int                    `json:"status"`
	Body   map[string]interface{} `json:"body"`
}

func ExecHttpRule(rule *rule.Rule) (types.RuleResponse, error) {
	if rule.Request.Http == nil {
		return nil, errors.New("rule does not have an http")
	}

	var bodyPayload io.Reader
	method := "GET"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	if rule.Request.Http.Body != nil {
		bodyData, err := json.Marshal(rule.Request.Http.Body)
		if err != nil {
			return nil, err
		}
		bodyPayload = strings.NewReader(string(bodyData))
	}

	client := &http.Client{Transport: tr}

	if rule.Request.Http.Method != nil {
		method = strings.ToUpper(*rule.Request.Http.Method)
	}

	req, err := http.NewRequest(method, rule.Request.Http.Url, bodyPayload)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if rule.Request.Http.Headers != nil {
		for key, value := range *rule.Request.Http.Headers {
			req.Header.Set(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var bodyMap map[string]interface{}
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		return nil, err
	}

	results := HttpResults{
		Status: resp.StatusCode,
		Body:   bodyMap,
	}

	result, err := structToMap(results)
	if err != nil {
		return nil, err
	}

	return result, nil
}
