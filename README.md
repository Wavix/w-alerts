# ElasticSearch Rule-Based Alert Engine

This application reads a directory containing rules (RULES_DIR) every minute, executes queries with data aggregation in ElasticSearch, and determines whether alerts should be triggered based on the conditions specified within these rules.

The application operates without a database. The current system status is available via the GET /status endpoint, which can be used by monitoring tools such as Zabbix. Additionally, the application supports bundling multiple rules in a single file and http requests.

### Example Response from GET /status:
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

- Flexible rule definitions supporting both HTTP requests and Elasticsearch queries.
- Real-time alerting and resolution with the current rule state accessible via the `/status` endpoint.
- Support for ratios, counts, and HTTP status checks, with the ability to combine multiple conditions.
- REST API endpoint (`GET /status`) providing the state of all configured rules.
- Compatibility with both aggregation-based and simple count-based Elasticsearch queries.
- Dynamic rule management in development mode, allowing rules to be added or edited without restarting the application.
- Operates without requiring a database, simplifying deployment and maintenance.
- Status page HTML interface for visual monitoring of alerts and system status.

## Rule settings

## Rule Configuration Options

The following table describes the configuration options available for defining a rule:

<table>
  <tr>
    <th>Option</th>
    <th>Description</th>
  </tr>
  <tr>
    <td><code>name</code></td>
    <td>The name of the rule. This will be displayed in the status response.</td>
  </tr>
  <tr>
    <td><code>description</code></td>
    <td>A description of the rule. This will be displayed in the status response.</td>
  </tr>
  <tr>
    <td><code>index</code></td>
    <td>The Elasticsearch index on which the query will be executed. The <code>*</code> character will be replaced with the current date.</td>
  </tr>
  <tr>
    <td><code>period</code></td>
    <td>The time period within which the rule's Elasticsearch query will be executed.</td>
  </tr>
  <tr>
    <td><code>interval</code></td>
    <td>The frequency at which the rule will be triggered.</td>
  </tr>
  <tr>
    <td><code>scope</code></td>
    <td>An optional attribute that will be used as a prefix in the rule's title.</td>
  </tr>
  <tr>
    <td><code>rules</code></td>
    <td>An array of conditions that must be met for the rule to trigger an alert.</td>
  </tr>
  <tr>
    <td><code>request</code></td>
    <td>The Elasticsearch or HTTP request associated with the rule.</td>
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

## Status Page HTML Interface

The application includes a visual status page that provides a real-time overview of all system alerts. This interface is accessible via the web browser and automatically updates to show the current state of all alerts.

### Features of the Status Page:

- Clean, responsive interface showing the current status of all alerts
- Separate section highlighting systems with issues
- Auto-refresh functionality with configurable intervals (10s, 30s, 1min, 5min)
- Visual indicators showing alert status (green for "ok", red for "problem")
- Displays alert names and descriptions for easy identification
- Manual refresh option for immediate status updates

### How to Access:

The status page is available at the root URL of the application (`/`) and serves the HTML interface from the `/public` directory.

### Auto-Deployment:

The status page HTML file is included in the release package along with the application binary when deployed using GitHub Actions.