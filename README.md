# GoCalc

[![Tests Status](https://github.com/m4tveevm/GoCalc/actions/workflows/go.yml/badge.svg)](https://github.com/m4tveevm/GoCalc/actions)
---
GoCalc (Calc is short for calculator)

GoCalc is small http server that could process some requests
in [Polish Notation](https://en.wikipedia.org/wiki/Polish_notation).
It has two main modules:

- The Calc module is designed to handle arithmetic expressions efficiently
  using the Shunting Yard Algorithm.
- Server module: Provides an HTTP server for processing requests.

_PS: There is also equivalent of such
algorithm [available in Python](https://github.com/m4tveevm/etu_algo_labs)._

## 🚀 How to Get Started

### Set up the server

1. Firstly you should build Docker-image:
   ```bash
   docker build -t gocalc .
   ```
2. And then just start GocCalc container:
   ```bash
   docker run -p 8080:8080 gocalc
   ```

### Examples of requests

#### Successful request (HTTP 200):

```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
    "expression": "2+2*2"
}'
```

That would return:

```json
{
   "result": 6
}
```

#### Invalid request (HTTP 422):

```bash
curl --location 'http://localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
    "expression": "2+2*0_0=>>>>52<<<<=0_0"
}'
```

That would return:

```json
{
   "error": "Expression is not valid"
}
```

#### But if something else goes wrong there would be such response (HTTP 500)

```json
{
   "error": "Internal server error"
}
```

## Built-In Tests

The GoCalc project includes unit tests for both the calculator logic and the
HTTP service.

1. To run tests you should use this commands:
   ```bash
   cd src
   go test ./... -v
   ```

Example output of successful tests:

```bash
=== RUN   TestCalculate
--- PASS: TestCalculate (0.00s)
PASS
ok      github.com/m4tveevm/GoCalc/calc (cached)
=== RUN   TestCalculateHandler
=== RUN   TestCalculateHandler/Valid_expression
=== RUN   TestCalculateHandler/Invalid_expression
=== RUN   TestCalculateHandler/Empty_expression
--- PASS: TestCalculateHandler (0.00s)
    --- PASS: TestCalculateHandler/Valid_expression (0.00s)
    --- PASS: TestCalculateHandler/Invalid_expression (0.00s)
    --- PASS: TestCalculateHandler/Empty_expression (0.00s)
PASS
ok      github.com/m4tveevm/GoCalc/server       (cached)
```

Each module's {name}_test.go file contains detailed tests for its
functionality.