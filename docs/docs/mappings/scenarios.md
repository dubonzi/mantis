# Scenarios

Scenarios are an *optional* feature that allow you to add state to especific mappings which will change how that mapping responds. They can be useful when writing tests that test how your application behaves when the state of your dependency changes.

Take the following as an example:

Your app depends on a `users` service which requires a valid token to access, and whenever that token expires, you expect your application to call the service to renew that token, you could mock this behaviour using the mappings below.

```json
{
  "scenario": {
    "name": "Renew Token",
    "startingState": true,
    "state": "Token Expired",
    "newState": "Renew"
  },
  "request": {
    "method": "GET",
    "path": {
      "exact": "/user-service/users/123",
    },
    "headers": {
      "Authorization": {
        "exact": "Bearer expired-token"
      }
    }
  },
  "response": {
    "statusCode": 403
  }
}
```

```json
{
  "scenario": {
    "name": "Renew Token",
    "state": "Renew",
    "newState": "Token Renewed"
  },
  "request": {
    "method": "POST",
    "path": {
      "exact": "/user-service/token",
    },
    "body": {
      "exact": "{\"clientId\": \"777\", \"secret\": \"secret-stuff\"}"
    }
  },
  "response": {
    "statusCode": 200,
    "body": "{\"token\": \"renewed-token\"}"
  }
}
```

```json
{
  "scenario": {
    "name": "Renew Token",
    "state": "Token Renewed",
  },
  "request": {
    "method": "GET",
    "path": {
      "exact": "/user-service/users/123",
    },
    "headers": {
      "Authorization": {
        "exact": "Bearer renewed-token"
      }
    }
  },
  "response": {
    "statusCode": 200,
    "body": "{\"id\": \"123\", \"name\": \"Frodo\", \"location\": \"The Shire\"}"
  }
}
```

When you first call `/user-service/users/123`, you'll get an 403 response, the scenario state will be updated to `Renew`, which means that the only request that will match this scenario now is the second mapping, the `POST` to renew the token. After that the state is set to `Token Renewed` and a `GET` with the renewed token will successfully return the user.

### Rules

There are a few rules for a scenario to be valid:

- A mapping will only be part of a scenario if the scenario name matches
- A scenario must have at least two states
- A scenario must have one, and only one, starting state
- States defined in `newState` must exist in the scenario
