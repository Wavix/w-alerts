# wAlerts

This is an application that reads a directory with rules (RULES_DIR) every minute, runs a query, including data aggregation in ElasticSearch, and decides whether an alert has occurred or not based on the conditions specified inside.

The application does not have a database. In the development environment, you can add, modify, and delete rules, and the app will re-read them within a minute. In production mode, the application reads the rules once upon startup.

The current registry state is available via GET /status for the zabbix agent (for ex.).

Also, the application supports a bundle of rules in one file

Example of a response to the GET /status request:

```json
{
  "status": [
    {
      "uuid": "8240a321-7dd6-ea42-39f6-da1a7f5deca9",
      "name": "Some rule name",
      "description": "Some description",
      "status": "ok" // or "problem"
    }
  ]
}
```

## Features

- flexible rules based on HTTP requests and Elasticsearch queries
- alert/resolve; Current rule state in the status
- ratio, count, status (http) requests and the ability to combine them
- the state of all rules is available via REST API GET /status
- aggregation queries and simple count queries
- adding and editing rules without the need for a restart in develop mode
- no database required

## Rule settings

<table>
    <tr>
        <th>Option</th>
        <th>Description</th>
    </tr>
    <tr>
        <td>name</td>
        <td>Problem name. It will be displayed in the status</td>
    </tr>
    <tr>
        <td>description</td>
        <td>Problem description. It will be displayed in the status</td>
    </tr>
    <tr>
        <td>index</td>
        <td>Index on which the ES query will be executed. * will be replaced with the current date</td>
    </tr>
    <tr>
        <td>period</td>
        <td>Period within which the rule (ES query) will be executed</td>
    </tr>
    <tr>
        <td>interval</td>
        <td>Rule triggering frequency</td>
    </tr>
    <tr>
        <td>rules</td>
        <td>Array of conditions under which the rule triggers an alert</td>
    </tr>
    <tr>
        <td>request</td>
        <td>ElasticSearch or Http request</td>
    </tr>
</table>

## Example of an ES query with aggregation

### Example of a condition for triggering a rule

```json
"rules": [
    {
      "field": "aggregations.total_requests",
      "operator": "gt",
      "value": 1000
    },
    {
      "field": "aggregations.error_requests",
      "field2": "aggregations.total_requests",
      "operator": "gt",
      "value": 0.5
    }
  ],
```
__In order for the rule to transition to the "problem" status, all conditions must be met__

In _field_, a nested key is specified for accessing the aggregated result.

If the condition contains the _field2_ field, it means that the condition turns into a ratio. That is, the result of the condition = field/field2.

The _field_ and _field2_ fields are only needed in the case of aggregated queries. For non-aggregated queries, the field fields do not need to be specified.

_Example of a query with aggregation:_

```json
"query": {
      "bool": {
        "must": [{ "term": { "url.keyword": "/v2/messages" } }]
      }
    },
    "aggs": {
      "total_requests": {
        "value_count": {
          "field": "status"
        }
      },
      "error_requests": {
        "filter": {
          "range": {
            "status": { "gt": 299 }
          }
        }
      }
    }
```

In this example, all aggregation values map to its key. That is:

```json
"total_requests": 123,
"error_requests": 1
```

## Example of an ES query witout aggregation

### Example of a condition for triggering a rule without aggregation

```json
"rules": [
    {
      "operator": "gt",
      "value": 1000
    }
  ],
```

_Example of a query without aggregation:_

```json
  "query": {
    "bool": {
      "must": [{ "term": { "url.keyword": "/v2/messages" } }]
    }
  }
```

## Example of a rule for monitoring an HTTP service

```json
{
  "name": "My super service is not responding",
  "description": "Error in the request to the super service public API. Status code: {}, country code: {}",
  "interval": "3m",
  "rules": [
    {
      "status": 200
    },
    {
      "field": "body.country_code",
      "operator": "eq",
      "value": "AR"
    }
  ],
  "request": {
    "http": {
      "url": "https://api.site.com/v1/validation?appid=XXX&type=format&phone_number=541152730593",
      "method": "GET"
    }
  }
}
```

If the status returned is not 200, or if the response attribute _body.country_code_ does not match "AR", the rule will transition to a _problem_ state.