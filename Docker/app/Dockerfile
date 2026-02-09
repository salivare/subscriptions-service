FROM golang:1.25.7-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o subscriptions ./cmd/subscriptions

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/subscriptions /app/subscriptions

# swagger.json + swaggerui
COPY --from=builder /app/swaggerui /app/swaggerui
COPY --from=builder /app/swagger.json /app/swagger.json

COPY --from=builder /app/docs.go /app/docs.go

ENTRYPOINT ["/app/subscriptions"]
