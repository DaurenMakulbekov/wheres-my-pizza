FROM golang:1.24.3-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o restaurant-system ./cmd

FROM alpine

COPY --from=builder /app/restaurant-system /restaurant-system

COPY --from=builder /app/configs/config.yaml /config.yaml

CMD ["./restaurant-system"]
