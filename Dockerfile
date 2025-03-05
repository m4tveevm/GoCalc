FROM golang:1.23.2 AS builder
WORKDIR /app

COPY ./src/go.mod ./
RUN go mod tidy

COPY ./src /app
RUN go build -o main ./server

FROM debian:bookworm
WORKDIR /root/

COPY --from=builder /app/main .


EXPOSE 8080
CMD ["./main"]
