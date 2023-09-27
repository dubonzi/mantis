# Response

The response object defines how and what that mapping will respond with when the request is matched.

```json
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
```

#### Body and Body File
> at least one required

You can set the body of the response directly in the mapping by using the `body` property or especify the path to a file which will be used as the response body. Note that the contents of the file will replace the `body` value.

### Delay
> optional

Mantis has support for defining a delay on the response for a mapping, that means Mantis will wait the especified time before returning the response everytime that mapping is matched with a request.

Note that very small durations (1ms or below) might not be accurate depending on how much load the application is experiencing or the network conditions.

#### Fixed Delay

Will delay the response by the especified time duration. 

```json
"delay": {
  "fixed": {
    "duration": "250ms"
  }
}
```