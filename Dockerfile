FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bot .

FROM alpine:3.19

RUN addgroup -S botgroup && adduser -S botuser -G botgroup

WORKDIR /app
COPY --from=builder /app/bot .

RUN chown botuser:botgroup /app/bot

USER botuser

CMD ["./bot"]