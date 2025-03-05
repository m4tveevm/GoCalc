FROM golang:1.23.2-alpine AS builder

WORKDIR /app
COPY . /app

WORKDIR /app/src
RUN go mod tidy

WORKDIR /app/src/orchestrator
RUN go build -o /orchestrator .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /orchestrator .
EXPOSE 8080
CMD ["./orchestrator"]
