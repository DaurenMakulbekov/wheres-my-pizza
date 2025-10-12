FROM golang:1.24.3-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o restaurant-system ./cmd

CMD ["./restaurant-system"]
