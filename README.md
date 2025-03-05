# GoCalc

### (Calc is short for calculator)

[![Tests Status](https://github.com/m4tveevm/GoCalc/actions/workflows/go.yml/badge.svg)](https://github.com/m4tveevm/GoCalc/actions)
---

#### GoCalc is a scalable, microservice-based calculator whose logic is based on [Polish Notation](https://en.wikipedia.org/wiki/Polish_notation).

It has three main modules:

- **Orchestrator** – handles incoming requests, assigns IDs to expressions, and
  maintains task statuses.
- **Agent** – periodically fetches tasks from the orchestrator, evaluates them
  using RPN calculation, and returns the results.
- **Calc** – a simple implementation from the previous task, responsible for
  evaluating mathematical expressions.

_PS: The original algorithm implementation is
also [available in Python](https://github.com/m4tveevm/etu_algo_labs)._

## 🚀 How to Get Started

### Start with Docker Compose

Ensure you have Docker and Docker Compose installed.

0. If you don't have Docker, download it from
   the [official site](https://www.docker.com) for your platform or use any
   docker-ready alternative (for example, [OrbStack](https://orbstack.dev)).
1. **Clone the repository** and navigate to the directory:

```bash
git clone https://github.com/m4tveevm/GoCalc.git
cd GoCalc
```

2. **Start the services** (orchestrator and agent):

```bash
docker-compose up --build
```

Services will be launched as follows:

- **Orchestrator** at `http://localhost:8080`
- **Agent** automatically connects to the orchestrator and begins processing
  tasks.

You can scale agents horizontally by running:

```bash
docker-compose up --scale agent=4
```

### Docker Compose example:

```yaml
services:
  orchestrator:
    build:
      context: .
      dockerfile: ./src/orchestrator/Dockerfile
    ports:
      - "8080:8080"
    environment:
      - TIME_ADDITION_MS=1000
      - TIME_SUBTRACTION_MS=1000
      - TIME_MULTIPLICATIONS_MS=1000
      - TIME_DIVISIONS_MS=1000

  agent:
    build:
      context: .
      dockerfile: ./src/agent/Dockerfile
    environment:
      - COMPUTING_POWER=2
      - ORCHESTRATOR_URL=http://orchestrator:8080
    depends_on:
      - orchestrator
```

### Examples of requests

#### Submitting an expression for calculation (HTTP `POST` request)

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
  "id": 1
}
```

#### Get calculation status by ID (HTTP `GET` request)

```bash
curl --location 'http://localhost:8080/api/v1/expressions/1'
```

That would return:
```json
{
  "expression": {
    "id": 1,
    "expression": "2+2*2",
    "status": "done",
    "result": 6
  }
}
```

#### Get all expressions (HTTP `GET` request)

```bash
curl --location 'http://localhost:8080/api/v1/expressions'
```

That would return:
```json
{
  "expressions": [
    {
      "id": 1,
      "expression": "2+2*2",
      "status": "done",
      "result": 6
    }
  ]
}
```

## Built-In Tests

The GoCalc project includes unit tests for each on of its modules:

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
--- PASS: TestCalculateHandler (0.00s)
PASS
ok      github.com/m4tveevm/GoCalc/orchestrator (cached)
```

> [!NOTE]
> A legacy version without microservice architecture is available on
> the [legacy branch](https://github.com/m4tveevm/GoCalc/tree/second-sprint).