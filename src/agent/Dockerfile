FROM golang:1.23.2-alpine AS builder

WORKDIR /app
COPY . /app

WORKDIR /app/src
RUN go mod tidy

WORKDIR /app/src/agent
RUN go build -o /agent .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /agent .
ENV COMPUTING_POWER=2
ENV ORCHESTRATOR_URL=http://orchestrator:8080
CMD ["./agent"]
