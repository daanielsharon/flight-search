FROM golang:1.24.1 AS builder

WORKDIR /app/shared
COPY shared .
RUN go mod download

WORKDIR /app/provider-service
COPY provider-service .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/provider-service .

CMD ["./main"]