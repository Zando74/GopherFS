
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main cmd/main.go


FROM alpine:latest

WORKDIR /app/

COPY --from=builder /app/main .
COPY --from=builder /app/config/config.yml .

ENV CONFIG_PATH=./config.yml

CMD ["./main"]