# syntax=docker/dockerfile:1.4
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache build-base zeromq-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go .
RUN go build -o server main.go

FROM alpine:latest
RUN apk --no-cache add zeromq
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]