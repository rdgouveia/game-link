FROM sqlc/sqlc:latest AS sqlc

FROM golang:1.25-alpine AS builder

COPY --from=sqlc /workspace/sqlc /usr/local/bin/sqlc

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN sqlc generate && go build -o server .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
