# Mappings

Mappings are the definitions Mantis uses to mock the responses to all requests, they contain contain information such as HTTP Method, path, headers, etc..

## Definition

### Basic 

The most basic mapping definition consists only of the request method, path and the response status code.

``` json
{
  "request": {
    "method": "GET",
    "path": {
      "exact": "/product/12345"
    }
  },
  "response": {
    "statusCode": 200
  }
}
```
This will match on any `GET` requests made to the `/products/12345` path and return a `200` status.

### Complete

Here is a complete example of a mapping with all it's fields:

``` json
{
  "scenario": {
    "name": "My Scenario",
    "startingState": true,
    "state": "First state",
    "newState": "Second state"
  },
  "request": {
    "method": "POST",
    "path": {
      "exact": "/products",
      "contains": ["product"],
      "pattern": ["/[a-zA-Z]+"]
    },
    "headers": {
      "Content-type": {
        "exact": "application/json"
      },
      "Authorization": {
        "contains": ["my-token"]
      },
      "Accept": {
        "pattern": ["json$", "xml$"]
      }
    },
     "body": {
      "exact": "{\"code\": \"12345\",\"name\": \"product\"}",
      "contains": ["product", "12345"],
      "pattern": ["\"code\": \"[0-9]+\""],
      "jsonPath": ["$[?(@.name == 'product')]"]
    },
  },
  "response": {
    "statusCode": 201,
    "headers": {
      "content-type": "application/json",
      "location": "/products/12345"
    },
    "body": "{\"id\": \"777\",\"code\": \"12345\",\"name\": \"product\"}",
    "bodyFile": "products/12345.json",
    "delay": {
      "fixed": {
        "duration": "250ms"
      }
    }
  }
}
```

As you can see, there are multiple ways of matching a certain component of the request. See [Matching](#matching) below for more information.
