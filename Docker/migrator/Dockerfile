FROM golang:1.25.7-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o migrator ./cmd/migrator

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/migrator /app/migrator
COPY migrations ./migrations

ENTRYPOINT ["/app/migrator"]
