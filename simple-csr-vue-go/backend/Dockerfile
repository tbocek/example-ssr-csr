FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod server.go ./
RUN go build -o server server.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]