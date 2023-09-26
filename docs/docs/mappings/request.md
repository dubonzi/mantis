# Request

The request object contains conditions about the request you want to match.

To match a request with a mapping, Mantis looks at each condition defined in the mapping and will only match it if all of them are true. Note that for any request component, only one type of condition can be used, that means you can't mix `exact` and `contains` to match the `path`.

The supported conditions are:

#### Exact 

> Works on Path, Headers and Body

Accepts only one value. Exact will compare the literal values and will be true if both are equal.

#### Contains

> Works on Path, Headers and Body

Accepts multiple values. Will be true if the value contains all of the especified strings.

For example, `"path": {"contains": ["products", "12345"]}` will match a request with path `/stores/products/12345`, but wont match `/stores/products/777`.

#### Regex

> Works on Path, Headers and Body

Accepts multiple values. Will be true if the value matches all of the especified patterns.

For example, `"path": {"pattern": ["/store/products/[0-9]+"]}` will match a request with path `/stores/products/12345`, but wont match `/stores/products/shoes`.

#### JSON Path

> Works on Body

Accepts multiple values. Will be true if the value matches all of the especified JSONPath patterns. Matching was implemented using [ojg](https://github.com/ohler55/ojg).

Example:

Mapping: 
```json
"body": {
  "jsonPath": ["$.products[?(@.id == '12345')]"]
}
```

Will match this request body: 

```json
{
  "products": [
    {"id": "12345"}, 
    {"id": "123452"}
  ]
}
```

