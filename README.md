# postback_delivery
A service to function as a small scale simulation of data distribution to third parties in real time.

## Prerequisites
Dependencies developed with, versions can probably change slightly:
* docker@1.12.3
* docker-compose@1.9.0
* Ubuntu 16.04.1 LTS

## Install
    docker-compose build
    docker-compose up

## Adding postback data
An example:

    curl localhost/ingest.php -H 'Content-Type: application/json' -d '{
          "endpoint":{
            "method":"GET",
            "url":"http://sample_domain_endpoint.com/data?title={mascot}&image={location}&foo={bar}"
          },
          "data":[
            {
              "mascot":"Gopher",
              "location":"https://blog.golang.org/gopher/gopher.png"
            }
          ]
        }

### Sample Request:
    (POST) http://{server_ip}/ingest.php
    (RAW POST DATA)
    {
      "endpoint":{
        "method":"GET",
        "url":"http://sample_domain_endpoint.com/data?title={mascot}&image={location}&foo={bar}"
      },
      "data":[
        {
          "mascot":"Gopher",
          "location":"https://blog.golang.org/gopher/gopher.png"
        }
      ]
    }

### Sample Response (Postback):
    GET http://sample_domain_endpoint.com/data?title=Gopher&image=https%3A%2F%2Fblog.golang.org%2Fgopher%2Fgopher.png&foo=

## Design requirements:
* Provision provided Ubuntu server (see Resources - Server) with software stack required to complete project.
* Change default redis port immediately on startup and add authentication
* Build a php application to ingest http requests, and a golang application to deliver http responses. Use Redis to host a job queue between them.

### Data flow:
* Web request (see sample request) >
* "Ingestion Agent" (php) >
* "Delivery Queue" (redis)
* "Delivery Agent" (go) >
* Web response (see sample response)

### App Operation - Ingestion Agent (php):
* Accept incoming http request
* Push a "postback" object to Redis for each "data" object contained in accepted request.

### Response
The Ingestion Agent responds with the following JSON data:

    {
        "redis_is_connected":"+PONG", // The response from a Redis ping. Empty if no connection.
        "postback_queue_size":0, // Queue size of all requests in the postback queue.
        "error":"MAX_QUEUE_SIZE <#> reached" // This key only exists if the MAX_QUEUE_SIZE is reached.
    }

Additionally, if `error` key exists, an `HTTP 503` will indicate to clients to try again later.

### App Operation - Delivery Agent (go):
* Continuously pull "postback" objects from Redis
* Deliver each postback object to http endpoint:
** Endpoint method: request.endpoint.method.
** Endpoint url: request.endpoint.url, with {xxx} replaced with values from each request.endpoint.data.xxx element.
* Log delivery time, response code, response time, and response body.

## Troubleshooting
For information on normal and error cases in the Delivery Agent, see logs under
`/var/log/`. `INFO` data is under `/var/log/main.root.INFO` and
`/var/log/main.root.ERROR`.

## Extra Merit
* (/) Clean, descriptive Git commit history.
* (/) Clean, easy-to-follow support documentation for an engineer attempting to troubleshoot your system.
* (/) All services should be configured to run automatically, and service should remain functional after system restarts.
* (/) High availability infrastructure considerations.
* Data integrity considerations, including safe shutdown.
* (/) Modular code design.
* Configurable default value for unmatched url {key}s.
* (/) Performance of system under external load.
* (/) Minimal bandwidth utilization between ingestion and delivery servers.
* Configurable response delivery retry attempts.
* Data validation / error handling.
* (/) Ability to deliver POST (as well as GET) responses.
* Service monitoring / application profiling.
* Delivery volume / success / failure visualizations.
* Internal benchmarking tool.
